package wechatbot

import (
	"strings"
	"testing"
)

func TestReadLimitedMediaRejectsOversizedResponse(t *testing.T) {
	_, err := readLimitedMedia(strings.NewReader("12345"), 4)
	if err == nil || !strings.Contains(err.Error(), "exceeds 4 bytes") {
		t.Fatalf("readLimitedMedia() error = %v", err)
	}
}
