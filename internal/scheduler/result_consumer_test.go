package scheduler

import (
	"context"
	"encoding/json"
	"testing"

	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
	"github.com/stretchr/testify/assert"
)

func TestProcessResult_UnknownType(t *testing.T) {
	// This test verifies that unknown result types are handled
	consumer := &ResultConsumer{}

	result := workerDto.WorkerResult{
		Type:   "unknown_type",
		TaskID: "test-task-1",
		Status: "completed",
	}

	ctx := context.Background()
	err := consumer.processResult(ctx, result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown result type")
}

func TestWorkerResultSerialization(t *testing.T) {
	result := workerDto.WorkerResult{
		Type:    workerDto.TaskTypeBackup,
		TaskID:  "test-task-1",
		JobID:   "test-job-1",
		Status:  "completed",
		Message: "Backup completed successfully",
		Data:    map[string]string{"path": "/backups/test"},
	}

	// Serialize
	data, err := json.Marshal(result)
	assert.NoError(t, err)

	// Deserialize
	var decoded workerDto.WorkerResult
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// Verify
	assert.Equal(t, result.Type, decoded.Type)
	assert.Equal(t, result.TaskID, decoded.TaskID)
	assert.Equal(t, result.JobID, decoded.JobID)
	assert.Equal(t, result.Status, decoded.Status)
	assert.Equal(t, result.Message, decoded.Message)
	assert.NotNil(t, decoded.Data)
}

func TestNewResultConsumer(t *testing.T) {
	consumer := NewResultConsumer(nil, "test_queue", nil, nil, nil, nil, nil)

	assert.NotNil(t, consumer)
	assert.Equal(t, "test_queue", consumer.queue)
}
