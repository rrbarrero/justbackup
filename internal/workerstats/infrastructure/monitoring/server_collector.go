package monitoring

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	"github.com/rrbarrero/justbackup/internal/workerstats/application"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// ServerStatsCollector collects and reports server resource statistics
type ServerStatsCollector struct {
	config  *config.ServerConfig
	service *application.WorkerStatsService
}

// NewServerStatsCollector creates a new server stats collector
func NewServerStatsCollector(cfg *config.ServerConfig, service *application.WorkerStatsService) *ServerStatsCollector {
	return &ServerStatsCollector{
		config:  cfg,
		service: service,
	}
}

// Start begins the collection loop
func (c *ServerStatsCollector) Start(ctx context.Context) {
	log.Printf("Monitoring started for backend server")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial report
	c.report(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.report(ctx)
		}
	}
}

func (c *ServerStatsCollector) report(ctx context.Context) {
	report, err := c.collect()
	if err != nil {
		log.Printf("Failed to collect backend stats: %v", err)
		return
	}

	if err := c.service.PostStats(ctx, report); err != nil {
		log.Printf("Failed to save backend stats: %v", err)
	}
}

func (c *ServerStatsCollector) collect() (application.WorkerStatsReportDTO, error) {
	report := application.WorkerStatsReportDTO{
		WorkerID:  "backend",
		Timestamp: time.Now(),
	}

	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return report, err
	}

	// CPU Usage (Process specific)
	cpuPct, err := p.CPUPercent()
	if err == nil {
		report.CPUUsage = cpuPct
	}

	// Memory (Process specific - RSS)
	memInfo, err := p.MemoryInfo()
	if err == nil {
		report.MemoryUsed = memInfo.RSS

		// System total for percentage calculation
		vm, err := mem.VirtualMemory()
		if err == nil {
			report.MemoryTotal = vm.Total
			report.MemoryPercent = (float64(report.MemoryUsed) / float64(vm.Total)) * 100
		}
	}

	// Disk (root backups path)
	// Using "/" or a specific path if available in config.
	// The server handles backups too if local storage is used.
	usage, err := disk.Usage("/")
	if err == nil {
		report.DiskTotal = usage.Total
		report.DiskUsed = usage.Used
		report.DiskPercent = usage.UsedPercent
	}

	return report, nil
}
