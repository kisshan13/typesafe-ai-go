package typesafe

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Image struct {
	data   []byte
	format string
}

func NewImage(data []byte, format string) Image {
	return Image{data: append([]byte(nil), data...), format: strings.ToLower(strings.TrimPrefix(format, "."))}
}

func ImageFromFile(path string) (Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Image{}, fmt.Errorf("typesafe: read image: %w", err)
	}
	format := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if format == "jpeg" {
		format = "jpg"
	}
	return NewImage(data, format), nil
}

type imageUpload struct {
	name   string
	ref    string
	data   []byte
	format string
}

func prepareEvaluationRequest(req EvaluationRequest) ([]byte, string, bool, error) {
	uploads := make(map[string]imageUpload)
	state, err := replaceImages(reflect.ValueOf(req.State), uploads)
	if err != nil {
		return nil, "", false, err
	}
	questions, err := replaceImages(reflect.ValueOf(req.Questions), uploads)
	if err != nil {
		return nil, "", false, err
	}
	assets := make(map[string]Asset, len(req.Assets)+len(uploads))
	for key, asset := range req.Assets {
		assets[key] = asset
	}
	for key, asset := range assets {
		if len(asset.Data) != 0 {
			if asset.Part == "" {
				asset.Part = "image-" + key
			}
			uploads[asset.Part] = imageUpload{name: asset.Part, data: asset.Data, format: asset.Format}
			asset.Type = "image"
			asset.Data = nil
			assets[key] = asset
		}
	}
	for _, upload := range uploads {
		if upload.ref != "" {
			assets[upload.ref] = Asset{Type: "image", Part: upload.name, Format: upload.format}
		}
	}
	if len(assets) == 0 {
		assets = nil
	}
	payload := map[string]any{"state": state, "model": req.Model, "questions": questions}
	if assets != nil {
		payload["assets"] = assets
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, "", false, fmt.Errorf("typesafe: encode request: %w", err)
	}
	if len(uploads) == 0 {
		return jsonBody, "application/json", len(assets) > 0 || hasBoxQuestion(questions), nil
	}
	body, contentType, err := encodeMultipart(jsonBody, uploads)
	return body, contentType, true, err
}

func encodeMultipart(request []byte, uploads map[string]imageUpload) ([]byte, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	requestHeader := make(textproto.MIMEHeader)
	requestHeader.Set("Content-Disposition", `form-data; name="request"`)
	requestHeader.Set("Content-Type", "application/json")
	part, err := writer.CreatePart(requestHeader)
	if err != nil {
		return nil, "", err
	}
	if _, err = part.Write(request); err != nil {
		return nil, "", err
	}
	for name, upload := range uploads {
		mimeType := "application/octet-stream"
		switch upload.format {
		case "png":
			mimeType = "image/png"
		case "jpg", "jpeg":
			mimeType = "image/jpeg"
		}
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, name, filepath.Base(name)))
		header.Set("Content-Type", mimeType)
		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, "", err
		}
		if _, err := part.Write(upload.data); err != nil {
			return nil, "", err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return body.Bytes(), writer.FormDataContentType(), nil
}

func replaceImages(value reflect.Value, uploads map[string]imageUpload) (any, error) {
	if !value.IsValid() {
		return nil, nil
	}
	if value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, nil
		}
		if value.CanInterface() {
			if image, ok := value.Interface().(Image); ok {
				return registerImage(image, uploads)
			}
			if image, ok := value.Interface().(*Image); ok && image != nil {
				return registerImage(*image, uploads)
			}
		}
		return replaceImages(value.Elem(), uploads)
	}
	switch value.Kind() {
	case reflect.Map:
		if value.IsNil() {
			return nil, nil
		}
		out := make(map[string]any, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			key := iter.Key()
			if key.Kind() != reflect.String {
				return nil, errors.New("typesafe: request maps must have string keys")
			}
			item, err := replaceImages(iter.Value(), uploads)
			if err != nil {
				return nil, err
			}
			out[key.String()] = item
		}
		return out, nil
	case reflect.Slice, reflect.Array:
		if value.Kind() == reflect.Slice && value.IsNil() {
			return nil, nil
		}
		out := make([]any, value.Len())
		for i := 0; i < value.Len(); i++ {
			item, err := replaceImages(value.Index(i), uploads)
			if err != nil {
				return nil, err
			}
			out[i] = item
		}
		return out, nil
	case reflect.Struct:
		if value.Type() == reflect.TypeOf(Image{}) {
			return registerImage(value.Interface().(Image), uploads)
		}
		out := make(map[string]any)
		typ := value.Type()
		for i := 0; i < value.NumField(); i++ {
			field := typ.Field(i)
			if !field.IsExported() || field.PkgPath != "" {
				continue
			}
			tag := strings.Split(field.Tag.Get("json"), ",")
			name := tag[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			}
			fv := value.Field(i)
			if len(tag) > 1 && strings.Contains(strings.Join(tag[1:], ","), "omitempty") && fv.IsZero() {
				continue
			}
			item, err := replaceImages(fv, uploads)
			if err != nil {
				return nil, err
			}
			out[name] = item
		}
		return out, nil
	default:
		return value.Interface(), nil
	}
}

func registerImage(image Image, uploads map[string]imageUpload) (string, error) {
	if len(image.data) == 0 {
		return "", errors.New("typesafe: image data cannot be empty")
	}
	if image.format != "png" && image.format != "jpg" && image.format != "jpeg" {
		return "", fmt.Errorf("typesafe: unsupported image format %q (want png, jpg, or jpeg)", image.format)
	}
	ref, err := newImageReference()
	if err != nil {
		return "", fmt.Errorf("typesafe: create image reference: %w", err)
	}
	part := "image-" + strings.ReplaceAll(ref, "-", "")
	uploads[part] = imageUpload{name: part, ref: ref, data: image.data, format: image.format}
	return ref, nil
}

func newImageReference() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}
	id[6] = id[6]&0x0f | 0x40
	id[8] = id[8]&0x3f | 0x80
	encoded := hex.EncodeToString(id[:])
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:], nil
}

func hasBoxQuestion(questions any) bool {
	questionMap, ok := questions.(map[string]any)
	if !ok {
		return false
	}
	for _, question := range questionMap {
		if fields, ok := question.(map[string]any); ok && fields["type"] == "box" {
			return true
		}
	}
	return false
}
