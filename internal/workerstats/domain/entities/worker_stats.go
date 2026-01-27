package entities

import (
	"time"
)

// WorkerStatsReport represents a single statistics report from a worker
type WorkerStatsReport struct {
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryTotal   uint64    `json:"memory_total"`
	MemoryUsed    uint64    `json:"memory_used"`
	MemoryPercent float64   `json:"memory_percent"`
	DiskTotal     uint64    `json:"disk_total"`
	DiskUsed      uint64    `json:"disk_used"`
	DiskPercent   float64   `json:"disk_percent"`
	Timestamp     time.Time `json:"timestamp"`
}

// WorkerStatsWindow maintains a rolling window of stats for a single worker
type WorkerStatsWindow struct {
	WorkerID string              `json:"worker_id"`
	Reports  []WorkerStatsReport `json:"reports"`
}

const MaxReportsPerWorker = 30

// NewWorkerStatsWindow creates a new window for a worker
func NewWorkerStatsWindow(workerID string) *WorkerStatsWindow {
	return &WorkerStatsWindow{
		WorkerID: workerID,
		Reports:  make([]WorkerStatsReport, 0, MaxReportsPerWorker),
	}
}

// AddReport adds a new report to the window, evicting the oldest if necessary
func (w *WorkerStatsWindow) AddReport(report WorkerStatsReport) {
	if len(w.Reports) >= MaxReportsPerWorker {
		w.Reports = w.Reports[1:]
	}
	w.Reports = append(w.Reports, report)
}
