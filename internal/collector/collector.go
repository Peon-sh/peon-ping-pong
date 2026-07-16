package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/peon-sh/peon-ping-pong/internal/config"
	"github.com/peon-sh/peon-ping-pong/pkg/api"
)

// Collector gathers host + container snapshots into a PushPayload.
type Collector struct {
	cfg    *config.Config
	host   *HostCollector
	docker *DockerCollector
	log    *slog.Logger
}

// New builds a Collector. docker may be nil when collector is disabled.
func New(cfg *config.Config, docker *DockerCollector, log *slog.Logger) *Collector {
	if log == nil {
		log = slog.Default()
	}
	return &Collector{
		cfg:    cfg,
		host:   NewHostCollector(),
		docker: docker,
		log:    log,
	}
}

// Collect builds a full push payload.
func (c *Collector) Collect(ctx context.Context) (api.PushPayload, error) {
	payload := api.PushPayload{
		SchemaVersion: api.SchemaVersion,
		ServerID:      c.cfg.ServerID,
		AgentVersion:  config.Version,
		SentAt:        time.Now().UTC(),
		Containers:    []api.Container{},
	}

	if !c.cfg.CollectorEnabled {
		return payload, nil
	}

	host, err := c.host.Collect(ctx)
	if err != nil {
		c.log.Warn("host collect failed", "err", err)
	} else {
		payload.Host = host
	}

	if c.docker != nil {
		containers, err := c.docker.Collect(ctx)
		if err != nil {
			c.log.Warn("docker collect failed", "err", err)
		} else {
			payload.Containers = containers
		}
	}

	return payload, nil
}
