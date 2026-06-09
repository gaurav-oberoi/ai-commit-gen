// Package llm talks to any OpenAI-compatible chat completions endpoint. The
// defaults point at a local Ollama instance so the tool runs offline with no
// API key, but setting OPENAI_BASE_URL / OPENAI_API_KEY lets it use a hosted
// provider instead.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultBaseURL = "http://localhost:11434/v1"
	defaultModel   = "llama3.2"
)

// Client is a minimal chat-completions caller.
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

// New builds a Client from explicit values, falling back to environment
// variables and then to the local Ollama defaults.
func New(baseURL, model string) *Client {
	if baseURL == "" {
		baseURL = env("OPENAI_BASE_URL", defaultBaseURL)
	}
	if model == "" {
		model = env("OPENAI_MODEL", defaultModel)
	}
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   model,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends a system + user prompt and returns the assistant reply.
func (c *Client) Complete(ctx context.Context, system, user string) (string, error) {
	payload := chatRequest{
		Model:       c.Model,
		Temperature: 0.2,
		Messages: []message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling %s: %w (is Ollama running? try `ollama serve`)", c.BaseURL, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("model endpoint returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("model error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("model returned no choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
