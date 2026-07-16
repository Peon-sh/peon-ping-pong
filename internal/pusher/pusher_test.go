package pusher

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/peon-sh/peon-ping-pong/pkg/api"
)

func TestPushOK(t *testing.T) {
	var gotAuth string
	var gotPayload api.PushPayload

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotPayload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))
	defer srv.Close()

	p := New(srv.URL, "test-token", slog.Default())
	err := p.Push(context.Background(), api.PushPayload{
		SchemaVersion: 1,
		ServerID:      "srv_1",
		AgentVersion:  "0.0.1",
		SentAt:        time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("auth = %q", gotAuth)
	}
	if gotPayload.ServerID != "srv_1" {
		t.Errorf("server_id = %q", gotPayload.ServerID)
	}
}

func TestPushUnauthorizedNoRetry(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	p := New(srv.URL, "bad", slog.Default())
	p.maxRetries = 3
	err := p.Push(context.Background(), api.PushPayload{ServerID: "s"})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestPushRetriesOn500(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := New(srv.URL, "tok", slog.Default())
	p.maxRetries = 3
	p.backoff = func(attempt int) time.Duration { return time.Millisecond }
	err := p.Push(context.Background(), api.PushPayload{ServerID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}
