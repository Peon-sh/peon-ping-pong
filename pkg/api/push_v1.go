// Package api defines the peon-ping-pong push contract (schema v1).
package api

import "time"

// SchemaVersion is the current push payload schema.
const SchemaVersion = 1

// PushPayload is the JSON body POSTed to PUSH_ENDPOINT.
type PushPayload struct {
	SchemaVersion int         `json:"schema_version"`
	ServerID      string      `json:"server_id"`
	AgentVersion  string      `json:"agent_version"`
	SentAt        time.Time   `json:"sent_at"`
	Host          HostMetrics `json:"host"`
	Containers    []Container `json:"containers"`
}

// HostMetrics is host-level resource usage.
type HostMetrics struct {
	CPUPercent      float64 `json:"cpu_percent"`
	MemoryPercent   float64 `json:"memory_percent"`
	DiskPercentRoot float64 `json:"disk_percent_root"`
	UptimeSeconds   uint64  `json:"uptime_seconds"`
}

// Container is a single Docker container snapshot.
type Container struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Health string `json:"health,omitempty"`
	Image  string `json:"image"`
}
