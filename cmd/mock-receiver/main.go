// Command mock-receiver is a tiny HTTP server that accepts peon-ping-pong pushes
// for local development and smoke tests.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/peon-sh/peon-ping-pong/pkg/api"
)

func main() {
	addr := envOr("LISTEN_ADDR", ":9090")
	expectToken := os.Getenv("TOKEN") // optional; if set, require Bearer match

	var mu sync.Mutex
	var last *api.PushPayload
	var count int

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /last", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if last == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"count": count,
			"last":  last,
		})
	})

	mux.HandleFunc("POST /api/v1/agents/push", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if expectToken != "" && auth != "Bearer "+expectToken {
			http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, `{"message":"bad body"}`, http.StatusBadRequest)
			return
		}
		var payload api.PushPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, `{"message":"invalid json"}`, http.StatusBadRequest)
			return
		}

		mu.Lock()
		last = &payload
		count++
		n := count
		mu.Unlock()

		log.Printf("push #%d server_id=%s agent=%s containers=%d cpu=%.1f mem=%.1f disk=%.1f",
			n, payload.ServerID, payload.AgentVersion, len(payload.Containers),
			payload.Host.CPUPercent, payload.Host.MemoryPercent, payload.Host.DiskPercentRoot)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	fmt.Printf("mock-receiver listening on %s\n", addr)
	fmt.Printf("  POST /api/v1/agents/push\n")
	fmt.Printf("  GET  /last\n")
	fmt.Printf("  GET  /health\n")
	log.Fatal(srv.ListenAndServe())
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
