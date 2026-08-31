package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	cfg.Workers = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestConfigJSONRoundTrip(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WebhookURL = "https://example.test/webhook"
	data, err := jsonMarshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	got, err := jsonUnmarshalConfig(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.WebhookURL != cfg.WebhookURL || got.BatchSize != cfg.BatchSize {
		t.Fatalf("round trip mismatch: %#v", got)
	}
}

func TestConfigPathUsesPlatformDirectory(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "config.json" {
		t.Fatalf("unexpected path: %s", path)
	}
}

func jsonMarshal(cfg Config) ([]byte, error) { return json.Marshal(cfg) }
func jsonUnmarshalConfig(b []byte) (Config, error) {
	var c Config
	err := json.Unmarshal(b, &c)
	return c, err
}

func TestSaveLoadResetConfig(t *testing.T) {
	old := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", old)
	tmp := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmp)
	cfg := DefaultConfig()
	cfg.WebhookURL = "secret"
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	got, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got.WebhookURL != "secret" {
		t.Fatalf("unexpected webhook: %q", got.WebhookURL)
	}
	if err := ResetConfig(); err != nil {
		t.Fatal(err)
	}
}
