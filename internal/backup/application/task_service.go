package application

import (
	"context"

	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
)

type BackupTaskService struct {
	publisher   interfaces.TaskPublisher
	resultStore interfaces.ResultStore
}

func NewBackupTaskService(publisher interfaces.TaskPublisher, resultStore interfaces.ResultStore) *BackupTaskService {
	return &BackupTaskService{
		publisher:   publisher,
		resultStore: resultStore,
	}
}

func (s *BackupTaskService) MeasureSize(ctx context.Context, hostID string, path string) (string, error) {
	return s.publisher.PublishMeasureTask(ctx, hostID, path)
}

func (s *BackupTaskService) GetTaskResult(ctx context.Context, taskID string) (string, error) {
	return s.resultStore.GetTaskResult(ctx, taskID)
}
