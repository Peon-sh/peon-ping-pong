# peon-ping-pong

Lightweight server monitoring agent for [Peon](https://peon.sh). Inspired by Coolify Sentinel, owned by Peon.

**Version:** `0.0.1`  
**Image:** `ghcr.io/peon-sh/peon-ping-pong:0.0.1`

Runs on each managed server, collects host + Docker container state, and **pushes** a JSON heartbeat to your Peon control plane over HTTPS.

> Peon app install/receiver integration is separate — this repo is the agent + contract only.

## Features (v0.0.1)

- Host metrics: CPU %, memory %, root disk %, uptime
- Docker containers: name, state, health, image
- Local HTTP: `GET /health`, `GET /version` on `:8888`
- Push loop with Bearer token auth + retry on 5xx / network errors
- Tiny `mock-receiver` for local smoke tests

## Quick start (Docker)

```bash
docker run -d \
  --name peon-ping-pong \
  --label peon.managed=true \
  -e TOKEN="your-token" \
  -e PUSH_ENDPOINT="https://app.peon.sh/api/v1/agents/push" \
  -e SERVER_ID="your-server-uuid" \
  -e PUSH_INTERVAL_SECONDS=60 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -p 127.0.0.1:8888:8888 \
  --health-cmd='wget -qO- http://127.0.0.1:8888/health || exit 1' \
  --health-interval=10s \
  --health-retries=3 \
  --restart unless-stopped \
  ghcr.io/peon-sh/peon-ping-pong:0.0.1
```

> Distroless image has no shell/`curl`; healthcheck via Docker’s HTTP probe or an external check is fine. For local Compose, use the included example.

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TOKEN` | yes | — | Bearer token for push auth |
| `PUSH_ENDPOINT` | yes | — | Full URL, e.g. `https://app.peon.sh/api/v1/agents/push` |
| `SERVER_ID` | yes | — | Peon server id / uuid |
| `PUSH_INTERVAL_SECONDS` | no | `60` | Push interval |
| `COLLECTOR_ENABLED` | no | `true` | Collect host + containers |
| `DEBUG` | no | `false` | Verbose JSON logs |
| `LISTEN_ADDR` | no | `:8888` | Local API bind address |
| `DOCKER_HOST` | no | `unix:///var/run/docker.sock` | Docker engine endpoint |

## Push payload (schema v1)

```json
{
  "schema_version": 1,
  "server_id": "cmrkpa4h900iy2kob16l5yhku",
  "agent_version": "0.0.1",
  "sent_at": "2026-07-16T10:00:00Z",
  "host": {
    "cpu_percent": 12.4,
    "memory_percent": 58.2,
    "disk_percent_root": 42.0,
    "uptime_seconds": 86400
  },
  "containers": [
    {
      "name": "peon-proxy",
      "state": "running",
      "health": "healthy",
      "image": "traefik:v3.6.6"
    }
  ]
}
```

`Authorization: Bearer <TOKEN>` must be set on every push.

## Local smoke test

```bash
# Terminal 1 — mock control plane
make build
TOKEN=dev ./bin/mock-receiver
# listens on :9090 — POST /api/v1/agents/push, GET /last

# Terminal 2 — agent
TOKEN=dev \
PUSH_ENDPOINT=http://127.0.0.1:9090/api/v1/agents/push \
SERVER_ID=local-dev \
PUSH_INTERVAL_SECONDS=10 \
./bin/peon-ping-pong

# Inspect last push
curl -s http://127.0.0.1:9090/last | jq .
curl -s http://127.0.0.1:8888/health
curl -s http://127.0.0.1:8888/version
```

Or with Compose:

```bash
docker compose -f docker-compose.example.yml up --build
```

## Develop

```bash
go test ./...
make build
```

Requires Go 1.25+.

## Release

Push a tag:

```bash
git tag v0.0.1
git push origin v0.0.1
```

GitHub Actions builds multi-arch (`amd64`, `arm64`) and publishes:

- `ghcr.io/peon-sh/peon-ping-pong:0.0.1`
- `ghcr.io/peon-sh/peon-ping-pong:latest`

## Layout

```
cmd/peon-ping-pong/   # agent entrypoint
cmd/mock-receiver/    # local push sink
internal/config/      # env config
internal/collector/   # host + docker
internal/pusher/      # HTTP push + retry
internal/server/      # /health /version
pkg/api/              # push schema types
```

## License

MIT
