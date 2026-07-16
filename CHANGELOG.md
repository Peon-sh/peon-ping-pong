# Changelog

## [0.0.1] — 2026-07-16

### Added

- Initial peon-ping-pong agent (Go)
- Host metrics collector (CPU, memory, disk, uptime)
- Docker container list collector
- HTTP pusher with Bearer auth and 5xx/network retries
- Local `/health` and `/version` API on `:8888`
- Push payload schema v1 (`pkg/api`)
- `mock-receiver` for local smoke tests
- Dockerfile (distroless), CI, and GHCR release workflow
- Image: `ghcr.io/peon-sh/peon-ping-pong:0.0.1`
