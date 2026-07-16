package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/peon-sh/peon-ping-pong/pkg/api"
)

// DockerClient lists containers (for tests).
type DockerClient interface {
	ContainerList(ctx context.Context) ([]dockerContainer, error)
	Close() error
}

type dockerContainer struct {
	Names  []string `json:"Names"`
	State  string   `json:"State"`
	Status string   `json:"Status"`
	Image  string   `json:"Image"`
}

// DockerCollector lists containers via the Docker Engine HTTP API.
type DockerCollector struct {
	cli DockerClient
}

type engineClient struct {
	http *http.Client
	base string
}

// NewDockerCollector connects to the Docker daemon over DOCKER_HOST.
func NewDockerCollector(dockerHost string) (*DockerCollector, error) {
	cli, err := newEngineClient(dockerHost)
	if err != nil {
		return nil, err
	}
	return &DockerCollector{cli: cli}, nil
}

// NewDockerCollectorWithClient is used in tests.
func NewDockerCollectorWithClient(cli DockerClient) *DockerCollector {
	return &DockerCollector{cli: cli}
}

// Close closes the Docker client.
func (d *DockerCollector) Close() error {
	if d.cli == nil {
		return nil
	}
	return d.cli.Close()
}

// Collect returns container snapshots (all containers, including stopped).
func (d *DockerCollector) Collect(ctx context.Context) ([]api.Container, error) {
	list, err := d.cli.ContainerList(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]api.Container, 0, len(list))
	for _, c := range list {
		health := ""
		switch {
		case strings.Contains(c.Status, "(healthy)"):
			health = "healthy"
		case strings.Contains(c.Status, "(unhealthy)"):
			health = "unhealthy"
		case strings.Contains(c.Status, "(starting)"):
			health = "starting"
		}
		out = append(out, api.Container{
			Name:   firstName(c.Names),
			State:  c.State,
			Health: health,
			Image:  c.Image,
		})
	}
	return out, nil
}

func firstName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.TrimPrefix(names[0], "/")
}

func newEngineClient(dockerHost string) (*engineClient, error) {
	if dockerHost == "" {
		dockerHost = "unix:///var/run/docker.sock"
	}

	var (
		network string
		addr    string
		base    string
	)

	switch {
	case strings.HasPrefix(dockerHost, "unix://"):
		network = "unix"
		addr = strings.TrimPrefix(dockerHost, "unix://")
		base = "http://localhost"
	case strings.HasPrefix(dockerHost, "tcp://"):
		network = "tcp"
		addr = strings.TrimPrefix(dockerHost, "tcp://")
		base = "http://" + addr
	case strings.HasPrefix(dockerHost, "http://") || strings.HasPrefix(dockerHost, "https://"):
		return &engineClient{
			http: &http.Client{Timeout: 15 * time.Second},
			base: strings.TrimRight(dockerHost, "/"),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported DOCKER_HOST: %s", dockerHost)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, network, addr)
		},
	}

	return &engineClient{
		http: &http.Client{Timeout: 15 * time.Second, Transport: transport},
		base: base,
	}, nil
}

func (e *engineClient) ContainerList(ctx context.Context) ([]dockerContainer, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.base+"/containers/json?all=true", nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("docker API status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var list []dockerContainer
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (e *engineClient) Close() error {
	e.http.CloseIdleConnections()
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
