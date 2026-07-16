package pusher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/peon-sh/peon-ping-pong/pkg/api"
)

// Pusher POSTs payloads to the control plane with retries.
type Pusher struct {
	endpoint   string
	token      string
	client     *http.Client
	log        *slog.Logger
	maxRetries int
	// backoff returns wait duration for a given attempt (1-based after first failure).
	backoff func(attempt int) time.Duration
}

// New creates a Pusher.
func New(endpoint, token string, log *slog.Logger) *Pusher {
	if log == nil {
		log = slog.Default()
	}
	return &Pusher{
		endpoint: endpoint,
		token:    token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		log:        log,
		maxRetries: 3,
		backoff: func(attempt int) time.Duration {
			return time.Duration(attempt*attempt) * time.Second
		},
	}
}

// Push sends one payload. Retries on network errors and 5xx.
func (p *Pusher) Push(ctx context.Context, payload api.PushPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			wait := p.backoff(attempt)
			p.log.Debug("retrying push", "attempt", attempt, "backoff", wait)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p.token)
		req.Header.Set("User-Agent", "peon-ping-pong/"+payload.AgentVersion)

		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = err
			p.log.Warn("push request failed", "err", err, "attempt", attempt)
			continue
		}

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			p.log.Debug("push ok", "status", resp.StatusCode)
			return nil
		}

		lastErr = fmt.Errorf("push failed: status=%d body=%s", resp.StatusCode, truncate(string(respBody), 200))
		// Do not retry auth / client errors (except 429).
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			p.log.Warn("push rejected", "status", resp.StatusCode, "attempt", attempt)
			continue
		}
		return lastErr
	}
	return lastErr
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
