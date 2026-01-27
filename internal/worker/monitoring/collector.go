package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// StatsCollector collects and reports worker statistics
type StatsCollector struct {
	config   *config.WorkerConfig
	workerID string
}

// WorkerStatsReportDTO matches the backend's expected DTO
type WorkerStatsReportDTO struct {
	WorkerID      string    `json:"worker_id"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryTotal   uint64    `json:"memory_total"`
	MemoryUsed    uint64    `json:"memory_used"`
	MemoryPercent float64   `json:"memory_percent"`
	DiskTotal     uint64    `json:"disk_total"`
	DiskUsed      uint64    `json:"disk_used"`
	DiskPercent   float64   `json:"disk_percent"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewStatsCollector creates a new stats collector with a fresh volatile ID
func NewStatsCollector(cfg *config.WorkerConfig) *StatsCollector {
	return &StatsCollector{
		config:   cfg,
		workerID: uuid.New().String(),
	}
}

// Start begins the collection loop
func (c *StatsCollector) Start(ctx context.Context) {
	log.Printf("Monitoring started for worker: %s", c.workerID)

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

func (c *StatsCollector) report(ctx context.Context) {
	report, err := c.collect()
	if err != nil {
		log.Printf("Failed to collect stats: %v", err)
		return
	}

	if err := c.send(ctx, report); err != nil {
		log.Printf("Failed to send stats to backend: %v", err)
	}
}

func (c *StatsCollector) collect() (WorkerStatsReportDTO, error) {
	report := WorkerStatsReportDTO{
		WorkerID:  c.workerID,
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

	// Memory (Process specific)
	memInfo, err := p.MemoryInfo()
	if err == nil {
		report.MemoryUsed = memInfo.RSS

		// We still need total system memory to calculate percentage of total
		vm, err := mem.VirtualMemory()
		if err == nil {
			report.MemoryTotal = vm.Total
			report.MemoryPercent = (float64(report.MemoryUsed) / float64(vm.Total)) * 100
		}
	}

	// Disk (backups path - this IS still system/partition wide)
	usage, err := disk.Usage(c.config.ContainerBackupRoot)
	if err == nil {
		report.DiskTotal = usage.Total
		report.DiskUsed = usage.Used
		report.DiskPercent = usage.UsedPercent
	}

	return report, nil
}

func (c *StatsCollector) send(ctx context.Context, report WorkerStatsReportDTO) error {
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/workers/stats", c.config.BackendURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("backend returned unexpected status: %s", resp.Status)
	}

	return nil
}
