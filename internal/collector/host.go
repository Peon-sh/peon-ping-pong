package collector

import (
	"context"
	"time"

	"github.com/peon-sh/peon-ping-pong/pkg/api"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// HostCollector gathers host CPU, memory, disk, and uptime.
type HostCollector struct{}

// NewHostCollector returns a HostCollector.
func NewHostCollector() *HostCollector {
	return &HostCollector{}
}

// Collect returns current host metrics.
func (h *HostCollector) Collect(ctx context.Context) (api.HostMetrics, error) {
	out := api.HostMetrics{}

	percents, err := cpu.PercentWithContext(ctx, 200*time.Millisecond, false)
	if err == nil && len(percents) > 0 {
		out.CPUPercent = round1(percents[0])
	}

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		out.MemoryPercent = round1(vm.UsedPercent)
	}

	du, err := disk.UsageWithContext(ctx, "/")
	if err == nil {
		out.DiskPercentRoot = round1(du.UsedPercent)
	}

	uptime, err := host.UptimeWithContext(ctx)
	if err == nil {
		out.UptimeSeconds = uptime
	}

	return out, nil
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
