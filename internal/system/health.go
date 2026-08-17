package system

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type DiskHealth struct {
	Path        string  `json:"path" jsonschema:"filesystem or mount path"`
	UsedPercent float64 `json:"used_percent" jsonschema:"percentage of disk space currently used"`
	FreeGB      float64 `json:"free_gb" jsonschema:"free disk space in gigabytes"`
}

type Health struct {
	OS            string       `json:"os" jsonschema:"operating system"`
	Architecture  string       `json:"architecture" jsonschema:"CPU architecture"`
	CPUPercent    float64      `json:"cpu_percent" jsonschema:"total CPU utilization percentage"`
	MemoryPercent float64      `json:"memory_percent" jsonschema:"memory utilization percentage"`
	MemoryFreeGB  float64      `json:"memory_free_gb" jsonschema:"available memory in gigabytes"`
	UptimeHours   float64      `json:"uptime_hours" jsonschema:"operating system uptime in hours"`
	Disks         []DiskHealth `json:"disks" jsonschema:"local filesystem usage"`
}

func GetHealth(ctx context.Context) (Health, error) {
	cpuPercent, err := cpu.PercentWithContext(ctx, 350*time.Millisecond, false)
	if err != nil { return Health{}, err }
	vm, err := mem.VirtualMemoryWithContext(ctx); if err != nil { return Health{}, err }
	uptime, err := host.UptimeWithContext(ctx); if err != nil { return Health{}, err }
	partitions, err := disk.PartitionsWithContext(ctx, false); if err != nil { return Health{}, err }
	disks := make([]DiskHealth, 0, len(partitions)); seen := map[string]bool{}
	for _, partition := range partitions {
		if seen[partition.Mountpoint] { continue }; seen[partition.Mountpoint] = true
		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint); if err != nil { continue }
		disks = append(disks, DiskHealth{Path: partition.Mountpoint, UsedPercent: round(usage.UsedPercent, 2), FreeGB: round(float64(usage.Free)/(1024*1024*1024), 2)})
	}
	totalCPU := 0.0; if len(cpuPercent) > 0 { totalCPU = cpuPercent[0] }
	return Health{OS: runtime.GOOS, Architecture: runtime.GOARCH, CPUPercent: round(totalCPU, 2), MemoryPercent: round(vm.UsedPercent, 2), MemoryFreeGB: round(float64(vm.Available)/(1024*1024*1024), 2), UptimeHours: round(float64(uptime)/3600, 2), Disks: disks}, nil
}

func round(value float64, places int) float64 { factor := 1.0; for i := 0; i < places; i++ { factor *= 10 }; return float64(int(value*factor+0.5)) / factor }
