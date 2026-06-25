package commands

import (
	"encoding/hex"
	"testing"

	"fkteams/internal/app/config"
)

func TestMaskString(t *testing.T) {
	tests := map[string]string{
		"":     "***",
		"a":    "***",
		"ab":   "***",
		"abc":  "a***c",
		"abcd": "a***d",
	}
	for input, want := range tests {
		if got := maskString(input); got != want {
			t.Fatalf("maskString(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestGenerateSecret(t *testing.T) {
	secret, err := generateSecret()
	if err != nil {
		t.Fatalf("generateSecret returned error: %v", err)
	}
	if len(secret) != 64 {
		t.Fatalf("secret length = %d, want 64", len(secret))
	}
	if _, err := hex.DecodeString(secret); err != nil {
		t.Fatalf("secret should be hex: %v", err)
	}
}

func TestEnableAndDisableAuth(t *testing.T) {
	useTempAppDir(t)
	cfg := config.Get()

	output := captureStdout(t, func() {
		if err := enableAuth(cfg, "admin", "secret"); err != nil {
			t.Fatalf("enableAuth returned error: %v", err)
		}
	})
	if !cfg.Server.Auth.Enabled || cfg.Server.Auth.Username != "admin" || cfg.Server.Auth.Password != "secret" {
		t.Fatalf("auth config after enable = %#v", cfg.Server.Auth)
	}
	if cfg.Server.Auth.Secret == "" {
		t.Fatal("enableAuth should generate missing secret")
	}
	if output == "" {
		t.Fatal("enableAuth should print status")
	}

	output = captureStdout(t, func() {
		if err := disableAuth(cfg); err != nil {
			t.Fatalf("disableAuth returned error: %v", err)
		}
	})
	if cfg.Server.Auth.Enabled {
		t.Fatalf("auth config after disable = %#v", cfg.Server.Auth)
	}
	if output == "" {
		t.Fatal("disableAuth should print status")
	}
}
