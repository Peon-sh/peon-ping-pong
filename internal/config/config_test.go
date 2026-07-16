package config

import (
	"testing"
)

func TestLoadMissingRequired(t *testing.T) {
	t.Setenv("TOKEN", "")
	t.Setenv("PUSH_ENDPOINT", "")
	t.Setenv("SERVER_ID", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing env")
	}
}

func TestLoadOK(t *testing.T) {
	t.Setenv("TOKEN", "secret")
	t.Setenv("PUSH_ENDPOINT", "https://app.peon.sh/api/v1/agents/push")
	t.Setenv("SERVER_ID", "srv_123")
	t.Setenv("PUSH_INTERVAL_SECONDS", "30")
	t.Setenv("COLLECTOR_ENABLED", "false")
	t.Setenv("DEBUG", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "secret" {
		t.Errorf("token = %q", cfg.Token)
	}
	if cfg.PushIntervalSeconds != 30 {
		t.Errorf("interval = %d", cfg.PushIntervalSeconds)
	}
	if cfg.CollectorEnabled {
		t.Error("expected collector disabled")
	}
	if !cfg.Debug {
		t.Error("expected debug true")
	}
}

func TestValidateBadEndpoint(t *testing.T) {
	cfg := &Config{
		Token:        "t",
		PushEndpoint: "ftp://bad",
		ServerID:     "s",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid endpoint error")
	}
}
