package protocol

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestReadLimitedBodyRejectsOversizedResponse(t *testing.T) {
	_, err := readLimitedBody(strings.NewReader("12345"), 4)
	if err == nil || !strings.Contains(err.Error(), "exceeds 4 bytes") {
		t.Fatalf("readLimitedBody() error = %v", err)
	}
}

func TestAPIPostPropagatesEncodingAndDecodingErrors(t *testing.T) {
	client := NewClient()
	if _, err := client.apiPost(context.Background(), "https://example.com", "/test", "token", make(chan int), time.Second); err == nil {
		t.Fatal("apiPost() should reject an unsupported request body")
	}

	client.HTTP = responseClient(http.StatusOK, "not-json")
	if _, err := client.apiPost(context.Background(), "https://example.com", "/test", "token", map[string]any{}, time.Second); err == nil {
		t.Fatal("apiPost() should reject malformed JSON")
	}
}

func TestAPIPostReturnsStructuredHTTPError(t *testing.T) {
	client := &Client{HTTP: responseClient(http.StatusUnauthorized, `{"ret":-14,"errmsg":"expired"}`)}
	_, err := client.apiPost(context.Background(), "https://example.com", "/test", "token", map[string]any{}, time.Second)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("apiPost() error = %v, want APIError", err)
	}
	if !apiErr.IsSessionExpired() || apiErr.HTTPStatus != http.StatusUnauthorized {
		t.Fatalf("apiPost() APIError = %+v", apiErr)
	}
}

func TestPollQRStatusRejectsMalformedResponse(t *testing.T) {
	client := &Client{HTTP: responseClient(http.StatusOK, "{")}
	_, err := client.PollQRStatus(context.Background(), "https://example.com", "qrcode")
	if err == nil {
		t.Fatal("PollQRStatus() should reject malformed JSON")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func responseClient(status int, body string) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}
}
