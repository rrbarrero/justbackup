package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerStatsWindow_AddReport(t *testing.T) {
	window := NewWorkerStatsWindow("worker-1")
	assert.Equal(t, 0, len(window.Reports))

	// Add 35 reports
	for i := 1; i <= 35; i++ {
		report := WorkerStatsReport{
			CPUUsage:  float64(i),
			Timestamp: time.Now(),
		}
		window.AddReport(report)
	}

	// Should only keep the last 30
	assert.Equal(t, 30, len(window.Reports))
	assert.Equal(t, float64(6), window.Reports[0].CPUUsage)
	assert.Equal(t, float64(35), window.Reports[29].CPUUsage)
}
