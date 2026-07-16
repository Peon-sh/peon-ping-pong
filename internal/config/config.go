package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Version is set at build time via -ldflags, defaults for local runs.
var Version = "0.0.1"

// Config holds runtime configuration from environment variables.
type Config struct {
	Token               string
	PushEndpoint        string
	PushIntervalSeconds int
	ServerID            string
	CollectorEnabled    bool
	Debug               bool
	ListenAddr          string
	DockerHost          string
}

// Load reads and validates configuration from the environment.
func Load() (*Config, error) {
	cfg := &Config{
		Token:               strings.TrimSpace(os.Getenv("TOKEN")),
		PushEndpoint:        strings.TrimSpace(os.Getenv("PUSH_ENDPOINT")),
		ServerID:            strings.TrimSpace(os.Getenv("SERVER_ID")),
		PushIntervalSeconds: 60,
		CollectorEnabled:    true,
		Debug:               false,
		ListenAddr:          ":8888",
		DockerHost:          envOr("DOCKER_HOST", "unix:///var/run/docker.sock"),
	}

	if v := os.Getenv("PUSH_INTERVAL_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("PUSH_INTERVAL_SECONDS must be a positive integer")
		}
		cfg.PushIntervalSeconds = n
	}

	if v := os.Getenv("COLLECTOR_ENABLED"); v != "" {
		cfg.CollectorEnabled = parseBool(v, true)
	}
	if v := os.Getenv("DEBUG"); v != "" {
		cfg.Debug = parseBool(v, false)
	}
	if v := os.Getenv("LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate ensures required fields are present.
func (c *Config) Validate() error {
	var missing []string
	if c.Token == "" {
		missing = append(missing, "TOKEN")
	}
	if c.PushEndpoint == "" {
		missing = append(missing, "PUSH_ENDPOINT")
	}
	if c.ServerID == "" {
		missing = append(missing, "SERVER_ID")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
	}
	if !strings.HasPrefix(c.PushEndpoint, "http://") && !strings.HasPrefix(c.PushEndpoint, "https://") {
		return fmt.Errorf("PUSH_ENDPOINT must be an http(s) URL")
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func parseBool(v string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
