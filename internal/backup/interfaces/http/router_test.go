package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application"
	"github.com/rrbarrero/justbackup/internal/backup/application/assembler"
	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/infrastructure/persistence/memory"
	backupHttp "github.com/rrbarrero/justbackup/internal/backup/interfaces/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRestoreRoute(t *testing.T) {
	backupRepo := memory.NewBackupRepositoryMemoryEmpty()
	backupErrorRepo := memory.NewBackupErrorRepositoryMemory()
	hostRepo := memory.NewHostRepositoryMemory()
	publisher := new(MockTaskPublisher)
	resultStore := new(MockResultStore)

	backupAssembler := assembler.NewBackupAssembler()
	hostService := application.NewHostService(hostRepo, backupRepo)

	lifecycleService := application.NewBackupLifecycleService(backupRepo, hostService, publisher, backupAssembler)
	queryService := application.NewBackupQueryService(backupRepo, hostService, backupErrorRepo, backupAssembler)
	searchService := application.NewBackupSearchService(backupRepo, hostService, nil, backupAssembler)
	restoreService := application.NewBackupRestoreService(backupRepo, hostService, publisher)
	taskService := application.NewBackupTaskService(publisher, resultStore)
	hookService := application.NewBackupHookService(backupRepo, backupAssembler)

	handler := backupHttp.NewBackupHandler(
		lifecycleService,
		queryService,
		searchService,
		restoreService,
		taskService,
		hookService,
	)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, func(hf http.HandlerFunc) http.HandlerFunc { return hf })

	// Create host and backup
	host := entities.NewHost("Test Host", "localhost", "user", 22, "path", false)
	_ = hostRepo.Save(context.TODO(), host)
	backup, err := entities.NewBackup(host.ID(), "/src", "/dest", entities.NewBackupSchedule("0 0 * * *"), []string{}, false, 0, false)
	assert.NoError(t, err)
	_ = backupRepo.Save(context.TODO(), backup)

	reqBody := dto.RestoreRequest{
		Path:         "/some/path",
		RestoreType:  "local",
		RestoreAddr:  "1.2.3.4:5678",
		RestoreToken: "secret",
	}
	body, _ := json.Marshal(reqBody)

	url := "/backups/" + backup.ID().String() + "/restore"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))

	expectedPath := "/mnt/backups/path/dest/some/path"
	publisher.On("PublishRestoreTask", mock.Anything, mock.Anything, expectedPath, reqBody.RestoreAddr, reqBody.RestoreToken).Return("task-123", nil)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code, "Route should match and return 202")
}
