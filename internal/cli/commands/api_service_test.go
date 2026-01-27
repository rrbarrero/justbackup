package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/cli/client"
	"github.com/rrbarrero/justbackup/internal/cli/config"
)

func TestAPIService_GetSSHKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/settings/ssh-key" {
			t.Errorf("expected path /settings/ssh-key, got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"publicKey": "ssh-rsa aaa"})
	}))
	defer server.Close()

	cfg := &config.Config{URL: server.URL}
	apiClient := client.NewClient(cfg)
	svc := NewAPIService(apiClient)

	key, err := svc.GetSSHKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "ssh-rsa aaa" {
		t.Errorf("expected ssh-rsa aaa, got %s", key)
	}
}

func TestAPIService_RegisterHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/hosts" {
			t.Errorf("expected POST /hosts, got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	cfg := &config.Config{URL: server.URL}
	apiClient := client.NewClient(cfg)
	svc := NewAPIService(apiClient)

	err := svc.RegisterHost(dto.CreateHostRequest{Name: "test-host"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAPIService_RequestRestore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/backups/b1/restore" {
			t.Errorf("expected POST /backups/b1/restore, got %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"task_id": "task-456"})
	}))
	defer server.Close()

	cfg := &config.Config{URL: server.URL}
	apiClient := client.NewClient(cfg)
	svc := NewAPIService(apiClient)

	taskID, err := svc.RequestRestore("b1", dto.RestoreRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if taskID != "task-456" {
		t.Errorf("expected task-456, got %s", taskID)
	}
}
