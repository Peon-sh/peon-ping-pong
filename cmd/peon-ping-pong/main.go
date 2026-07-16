package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/peon-sh/peon-ping-pong/internal/collector"
	"github.com/peon-sh/peon-ping-pong/internal/config"
	"github.com/peon-sh/peon-ping-pong/internal/pusher"
	"github.com/peon-sh/peon-ping-pong/internal/server"
)

func main() {
	level := slog.LevelInfo
	if os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "1" {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config error", "err", err)
		os.Exit(1)
	}

	log.Info("starting peon-ping-pong",
		"version", config.Version,
		"server_id", cfg.ServerID,
		"push_interval_seconds", cfg.PushIntervalSeconds,
		"collector_enabled", cfg.CollectorEnabled,
	)

	var dockerCol *collector.DockerCollector
	if cfg.CollectorEnabled {
		dockerCol, err = collector.NewDockerCollector(cfg.DockerHost)
		if err != nil {
			log.Warn("docker client unavailable; container metrics disabled", "err", err)
		} else {
			defer dockerCol.Close()
		}
	}

	col := collector.New(cfg, dockerCol, log)
	p := pusher.New(cfg.PushEndpoint, cfg.Token, log)
	srv := server.New(cfg.ListenAddr, config.Version, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("local API stopped", "err", err)
			stop()
		}
	}()

	// Immediate first push, then on interval.
	runPush(ctx, col, p, log)

	ticker := time.NewTicker(time.Duration(cfg.PushIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("shutting down")
			return
		case <-ticker.C:
			runPush(ctx, col, p, log)
		}
	}
}

func runPush(ctx context.Context, col *collector.Collector, p *pusher.Pusher, log *slog.Logger) {
	payload, err := col.Collect(ctx)
	if err != nil {
		log.Error("collect failed", "err", err)
		return
	}
	if err := p.Push(ctx, payload); err != nil {
		log.Error("push failed", "err", err)
		return
	}
	log.Info("push ok",
		"containers", len(payload.Containers),
		"cpu_percent", payload.Host.CPUPercent,
		"memory_percent", payload.Host.MemoryPercent,
		"disk_percent_root", payload.Host.DiskPercentRoot,
	)
}
