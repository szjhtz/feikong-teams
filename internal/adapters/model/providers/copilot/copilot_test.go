package copilot

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAgentInitiatorContext(t *testing.T) {
	ctx := context.Background()
	if isAgentInitiator(ctx) {
		t.Fatal("background context should not be an agent initiator")
	}
	if !isAgentInitiator(WithAgentInitiator(ctx)) {
		t.Fatal("WithAgentInitiator context should be an agent initiator")
	}
}

func TestDetectInitiator(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "invalid json", body: "{", want: "user"},
		{name: "empty messages", body: `{"messages":[]}`, want: "user"},
		{name: "last user", body: `{"messages":[{"role":"assistant"},{"role":"user"}]}`, want: "user"},
		{name: "last assistant", body: `{"messages":[{"role":"user"},{"role":"assistant"}]}`, want: "agent"},
		{name: "last tool", body: `{"messages":[{"role":"tool"}]}`, want: "agent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectInitiator([]byte(tt.body)); got != tt.want {
				t.Fatalf("detectInitiator(%s) = %q, want %q", tt.body, got, tt.want)
			}
		})
	}
}

func TestDetectVision(t *testing.T) {
	if !detectVision([]byte(`{"content":[{"type":"image_url"}]}`)) {
		t.Fatal("compact image_url content should be detected")
	}
	if !detectVision([]byte(`{"content":[{"type": "image_url"}]}`)) {
		t.Fatal("spaced image_url content should be detected")
	}
	if detectVision([]byte(`{"content":[{"type":"text"}]}`)) {
		t.Fatal("text content should not be detected as vision")
	}
}

func TestCopilotTransportInjectsHeaders(t *testing.T) {
	t.Setenv("FEIKONG_APP_DIR", t.TempDir())
	tm := &TokenManager{token: &Token{
		GitHubToken:  "gh",
		CopilotToken: "copilot",
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
	}}
	var captured *http.Request
	transport := &copilotTransport{
		tm: tm,
		base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			captured = req.Clone(req.Context())
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			captured.Body = io.NopCloser(strings.NewReader(string(body)))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     http.Header{},
				Request:    req,
			}, nil
		}),
	}
	body := `{"messages":[{"role":"assistant"}],"content":[{"type":"image_url"}]}`
	req, err := http.NewRequest("POST", "https://example.com/chat", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("X-Api-Key", "remove-me")

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	resp.Body.Close()

	if captured.Header.Get("Authorization") != "Bearer copilot" {
		t.Fatalf("Authorization = %q, want bearer token", captured.Header.Get("Authorization"))
	}
	if captured.Header.Get("X-Initiator") != "agent" {
		t.Fatalf("X-Initiator = %q, want agent", captured.Header.Get("X-Initiator"))
	}
	if captured.Header.Get("Openai-Intent") != "conversation-agent" {
		t.Fatalf("Openai-Intent = %q", captured.Header.Get("Openai-Intent"))
	}
	if captured.Header.Get("X-Interaction-Type") != "conversation-agent" {
		t.Fatalf("X-Interaction-Type = %q", captured.Header.Get("X-Interaction-Type"))
	}
	if captured.Header.Get("Copilot-Vision-Request") != "true" {
		t.Fatalf("Copilot-Vision-Request = %q, want true", captured.Header.Get("Copilot-Vision-Request"))
	}
	if captured.Header.Get("X-Api-Key") != "" {
		t.Fatalf("X-Api-Key = %q, want removed", captured.Header.Get("X-Api-Key"))
	}
	for _, header := range []string{"User-Agent", "Editor-Version", "Editor-Plugin-Version", "Copilot-Integration-Id", "X-Github-Api-Version", "X-Request-Id", "X-Agent-Task-Id", "Vscode-Sessionid", "Editor-Device-Id", "X-Interaction-Id"} {
		if captured.Header.Get(header) == "" {
			t.Fatalf("%s header is empty", header)
		}
	}
}

func TestCopilotTransportUsesAgentInitiatorContext(t *testing.T) {
	t.Setenv("FEIKONG_APP_DIR", t.TempDir())
	tm := &TokenManager{token: &Token{
		GitHubToken:  "gh",
		CopilotToken: "copilot",
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
	}}
	var initiator string
	transport := &copilotTransport{
		tm: tm,
		base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			initiator = req.Header.Get("X-Initiator")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     http.Header{},
				Request:    req,
			}, nil
		}),
	}
	req, err := http.NewRequestWithContext(WithAgentInitiator(context.Background()), "POST", "https://example.com/chat", strings.NewReader(`{"messages":[{"role":"user"}]}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	resp.Body.Close()
	if initiator != "agent" {
		t.Fatalf("X-Initiator = %q, want agent from context", initiator)
	}
}

func TestReadRequestBodyRejectsOversizedBody(t *testing.T) {
	if _, err := readRequestBody(strings.NewReader("12345"), 4); err == nil {
		t.Fatal("oversized Copilot request body was accepted")
	}
}

func TestSetReplayableRequestBodyRestoresContent(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	setReplayableRequestBody(req, []byte("payload"))
	first, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := req.GetBody()
	if err != nil {
		t.Fatal(err)
	}
	defer replay.Close()
	second, err := io.ReadAll(replay)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != "payload" || string(second) != "payload" || req.ContentLength != 7 {
		t.Fatalf("replayed bodies = %q, %q; length = %d", first, second, req.ContentLength)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
