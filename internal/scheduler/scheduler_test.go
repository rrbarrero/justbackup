package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/rrbarrero/justbackup/internal/backup/infrastructure/persistence/memory"
	"github.com/stretchr/testify/assert"
)

func TestScheduler_ProcessDueBackups(t *testing.T) {
	// Setup
	backupRepo := memory.NewBackupRepositoryMemoryEmpty()
	redisPublisher := &RedisPublisher{} // We'll need a mock or test instance

	// Create a scheduler
	_ = NewScheduler(backupRepo, nil, redisPublisher, 1*time.Minute)

	// This test would require a real Redis connection and the Backup entity
	// to have SetNextRunAt method. Skipping for now.
	// In a real test, we'd use a mock Redis or test container
	t.Skip("Requires Redis mock and API changes")
}

func TestScheduler_Start_Context_Cancellation(t *testing.T) {
	// Setup
	backupRepo := memory.NewBackupRepositoryMemoryEmpty()
	redisPublisher := &RedisPublisher{} // Mock
	scheduler := NewScheduler(backupRepo, nil, redisPublisher, 100*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start scheduler in goroutine
	done := make(chan bool)
	go func() {
		scheduler.Start(ctx)
		done <- true
	}()

	// Wait for completion
	select {
	case <-done:
		// Scheduler stopped gracefully
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Scheduler did not stop within timeout")
	}
}

func TestNewScheduler(t *testing.T) {
	backupRepo := memory.NewBackupRepositoryMemoryEmpty()
	redisPublisher := &RedisPublisher{}
	interval := 1 * time.Minute

	scheduler := NewScheduler(backupRepo, nil, redisPublisher, interval)

	assert.NotNil(t, scheduler)
	assert.Equal(t, interval, scheduler.interval)
	assert.NotNil(t, scheduler.repo)
	assert.NotNil(t, scheduler.publisher)
}
