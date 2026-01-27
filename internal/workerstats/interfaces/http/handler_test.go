package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/websocket"
	"github.com/rrbarrero/justbackup/internal/workerstats/application"
	"github.com/rrbarrero/justbackup/internal/workerstats/domain/entities"
	"github.com/rrbarrero/justbackup/internal/workerstats/infrastructure/persistence/memory"
	"github.com/stretchr/testify/assert"
)

func TestWorkerStatsHandler_PostStats(t *testing.T) {
	repo := memory.NewWorkerStatsRepositoryMemory()
	service := application.NewWorkerStatsService(repo, websocket.NewHub())
	handler := NewWorkerStatsHandler(service)

	report := application.WorkerStatsReportDTO{
		WorkerID: "test-worker",
		CPUUsage: 10.5,
	}
	body, _ := json.Marshal(report)

	req, _ := http.NewRequest("POST", "/workers/stats", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.PostStats(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)

	// Verify repo has the data
	stats, _ := repo.GetStats(context.Background(), "test-worker")
	assert.NotNil(t, stats)
	assert.Equal(t, 1, len(stats.Reports))
	assert.Equal(t, 10.5, stats.Reports[0].CPUUsage)
}

func TestWorkerStatsHandler_GetStats(t *testing.T) {
	repo := memory.NewWorkerStatsRepositoryMemory()
	service := application.NewWorkerStatsService(repo, websocket.NewHub())
	handler := NewWorkerStatsHandler(service)

	// Pre-seed some data
	err := repo.SaveReport(context.Background(), "worker-1", entities.WorkerStatsReport{CPUUsage: 5.0})
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/workers/stats", nil)
	rr := httptest.NewRecorder()

	handler.GetStats(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result []*entities.WorkerStatsWindow
	err = json.NewDecoder(rr.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "worker-1", result[0].WorkerID)
}
