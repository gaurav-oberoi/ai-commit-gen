package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompleteHappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "hello") {
			t.Errorf("user prompt not forwarded: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "  hi there  "}},
			},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "test-model")
	got, err := c.Complete(context.Background(), "sys", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hi there" {
		t.Fatalf("expected trimmed content, got %q", got)
	}
}

func TestCompleteErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "model exploded")
	}))
	defer srv.Close()

	c := New(srv.URL, "m")
	if _, err := c.Complete(context.Background(), "s", "u"); err == nil {
		t.Fatal("expected error on 500 status")
	}
}

func TestNewUsesDefaults(t *testing.T) {
	t.Setenv("OPENAI_BASE_URL", "")
	t.Setenv("OPENAI_MODEL", "")
	c := New("", "")
	if c.BaseURL != defaultBaseURL {
		t.Fatalf("expected default base url, got %s", c.BaseURL)
	}
	if c.Model != defaultModel {
		t.Fatalf("expected default model, got %s", c.Model)
	}
}
