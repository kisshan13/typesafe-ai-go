package typesafe

import (
	"context"
	"net/http"
	"net/http/httptest"
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
