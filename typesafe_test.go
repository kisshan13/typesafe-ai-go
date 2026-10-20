package typesafe

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEvaluate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/systemone" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Errorf("authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"model":"jev-1.13.0","answers":{"urgent":{"type":"noul","noul":0.95},"team":{"type":"choice","choice":"billing","probabilities":{"billing":0.8,"technical":0.2},"confidence":0.7}},"usage":{"input_tokens":10,"output_tokens":4}}`))
	}))
	defer server.Close()

	client := NewClient("secret", WithBaseURL(server.URL))
	result, err := client.Evaluate(context.Background(), EvaluationRequest{
		State: "A payout failed", Model: DefaultModel,
		Questions: map[string]Question{
			"urgent": Noul("Is this urgent?", nil),
			"team":   Choice("Which team?", map[string]any{"billing": "payments", "technical": nil}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	urgent, ok := result.Answers.Get("urgent")
	if !ok || urgent.Noul != 0.95 {
		t.Fatalf("unexpected urgent answer: %#v", result.Answers)
	}
	if result.Answers.MustGet("team").Choice != "billing" {
		t.Fatalf("unexpected answers: %#v", result.Answers)
	}
}

func TestEvaluateRetries429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{"model":"jev-latest","answers":{},"usage":{}}`))
	}))
	defer server.Close()
	client := NewClient("key", WithBaseURL(server.URL), WithMaxRetries(1), WithRetryBackoff(time.Millisecond))
	_, err := client.Evaluate(context.Background(), EvaluationRequest{State: "x", Model: DefaultModel, Questions: map[string]Question{"q": Noul("?", nil)}})
	if err != nil || attempts != 2 {
		t.Fatalf("err=%v attempts=%d", err, attempts)
	}
}

func TestEvaluateValidation(t *testing.T) {
	_, err := NewClient("key").Evaluate(context.Background(), EvaluationRequest{Model: DefaultModel})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestDefaults(t *testing.T) {
	client := NewClient("key", WithBaseURL(""))
	if client.baseURL != DefaultBaseURL {
		t.Fatalf("base URL = %q, want %q", client.baseURL, DefaultBaseURL)
	}
	if JevLatest != "jev-latest" || DefaultModel != JevLatest {
		t.Fatalf("unexpected model defaults: %q, %q", JevLatest, DefaultModel)
	}
}

func TestEvaluationRequestBuilder(t *testing.T) {
	req := NewEvaluationRequest("state").
		Noul("urgent", "Is this urgent?", nil).
		Choice("team", "Which team?", map[string]any{"billing": nil}).
		Build()
	if req.Model != JevLatest || len(req.Questions) != 2 {
		t.Fatalf("unexpected request: %#v", req)
	}
}

func TestEvaluateImageUpload(t *testing.T) {
	imageData := []byte("test image bytes")
	if path := os.Getenv("TYPESAFE_TEST_IMAGE"); path != "" {
		var err error
		imageData, err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
	}
	image := NewImage(imageData, "jpg")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("TypeSafe-Danger-Zone-Images"); got != "true" {
			t.Errorf("image header = %q", got)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "multipart/form-data;") {
			t.Errorf("content type = %q", got)
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Errorf("multipart reader: %v", err)
			return
		}
		requestPart, err := reader.NextPart()
		if err != nil {
			t.Errorf("request part: %v", err)
			return
		}
		if requestPart.FormName() != "request" {
			t.Errorf("first part name = %q, want request", requestPart.FormName())
		}
		var body struct {
			State  map[string]string `json:"state"`
			Assets map[string]Asset `json:"assets"`
		}
		if err := json.NewDecoder(requestPart).Decode(&body); err != nil {
			t.Errorf("decode request part: %v", err)
			return
		}
		ref := body.State["photo"]
		asset, ok := body.Assets[ref]
		if !ok || asset.Type != "image" || asset.Format != "jpg" {
			t.Errorf("image asset = %#v, found=%v", asset, ok)
		}
		imagePart, err := reader.NextPart()
		if err != nil {
			t.Errorf("image part: %v", err)
			return
		}
		if imagePart.FormName() != asset.Part {
			t.Errorf("image part name = %q, asset part = %q", imagePart.FormName(), asset.Part)
		}
		got, err := io.ReadAll(imagePart)
		if err != nil {
			t.Errorf("read image part: %v", err)
		} else if !bytes.Equal(got, imageData) {
			t.Errorf("uploaded image bytes do not match")
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"model":"jev-1.14.0","answers":{"mug":{"type":"box","box":{"top":0.25,"bottom":0.82,"left":0.44,"right":0.84},"confidence":0.92}},"usage":{"input_tokens":290,"output_tokens":21}}`)
	}))
	defer server.Close()

	client := NewClient("secret", WithBaseURL(server.URL), WithMaxRetries(0))
	request := NewEvaluationRequest(map[string]any{"photo": image}).
		Model(JevPreview).
		Box("mug", "the coffee mug in `photo`").
		Build()
	result, err := client.Evaluate(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	box := result.Answers.MustGet("mug").Box
	if box == nil || box.Top != 0.25 || box.Right != 0.84 {
		t.Fatalf("unexpected box answer: %#v", box)
	}
}
