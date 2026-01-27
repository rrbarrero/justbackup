package commands

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
)

func TestRestoreService_ExecuteRemote(t *testing.T) {
	var capturedReq dto.RestoreRequest
	apiMock := &MockAPIService{
		RequestRestoreFunc: func(backupID string, req dto.RestoreRequest) (string, error) {
			capturedReq = req
			return "task-123", nil
		},
	}
	netMock := &MockNetService{}
	svc := NewRestoreService(apiMock, netMock)

	params := RemoteRestoreParams{
		BackupID:     "b1",
		Path:         "/var/log",
		TargetHostID: "h1",
		TargetPath:   "/tmp",
	}

	taskID, err := svc.ExecuteRemote(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if taskID != "task-123" {
		t.Fatalf("unexpected taskID: %s", taskID)
	}

	if capturedReq.RestoreType != "remote" || capturedReq.TargetPath != "/tmp" || capturedReq.TargetHostID != "h1" {
		t.Errorf("unexpected request parameters: %+v", capturedReq)
	}
}

func TestRestoreService_ExecuteLocal(t *testing.T) {
	var capturedReq dto.RestoreRequest
	apiMock := &MockAPIService{
		RequestRestoreFunc: func(backupID string, req dto.RestoreRequest) (string, error) {
			capturedReq = req
			return "", nil
		},
	}

	var capturedToken string
	netMock := &MockNetService{
		ListenTCPFunc: func() (string, int, io.Closer, error) {
			return "127.0.0.1", 9999, &mockCloser{}, nil
		},
		AcceptAndValidateFunc: func(listener io.Closer, token string) (io.ReadCloser, error) {
			capturedToken = token
			return io.NopCloser(strings.NewReader("")), nil
		},
	}
	svc := NewRestoreService(apiMock, netMock)

	params := LocalRestoreParams{
		BackupID:  "b1",
		Path:      "/var/log",
		LocalDest: "/tmp/restore",
	}

	err := svc.ExecuteLocal(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.RestoreType != "local" || capturedReq.RestoreAddr != "127.0.0.1:9999" {
		t.Errorf("unexpected request parameters: %+v", capturedReq)
	}

	if capturedReq.RestoreToken == "" {
		t.Error("expected token to be generated and sent")
	}

	if capturedToken != capturedReq.RestoreToken {
		t.Errorf("token mismatch: accepted %s, but sent %s", capturedToken, capturedReq.RestoreToken)
	}
}

func TestRestoreService_ExecuteLocal_CustomAddr(t *testing.T) {
	apiMock := &MockAPIService{}
	var capturedAddr string
	apiMock.RequestRestoreFunc = func(backupID string, req dto.RestoreRequest) (string, error) {
		capturedAddr = req.RestoreAddr
		return "", nil
	}

	netMock := &MockNetService{
		ListenTCPFunc: func() (string, int, io.Closer, error) {
			return "127.0.0.1", 9999, &mockCloser{}, nil
		},
	}
	svc := NewRestoreService(apiMock, netMock)

	params := LocalRestoreParams{
		BackupID:   "b1",
		CustomAddr: "192.168.1.50",
	}

	_ = svc.ExecuteLocal(params)
	if capturedAddr != "192.168.1.50:9999" {
		t.Errorf("expected 192.168.1.50:9999, got %s", capturedAddr)
	}
}

func TestRestoreService_ErrorPropagation(t *testing.T) {
	apiMock := &MockAPIService{
		RequestRestoreFunc: func(backupID string, req dto.RestoreRequest) (string, error) {
			return "", fmt.Errorf("api error")
		},
	}
	netMock := &MockNetService{}
	svc := NewRestoreService(apiMock, netMock)

	err := svc.ExecuteLocal(LocalRestoreParams{BackupID: "b1"})
	if err == nil || !strings.Contains(err.Error(), "api error") {
		t.Errorf("expected api error, got %v", err)
	}
}
