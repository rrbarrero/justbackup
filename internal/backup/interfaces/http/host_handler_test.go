package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application"
	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/infrastructure/persistence/memory"
	backupHttp "github.com/rrbarrero/justbackup/internal/backup/interfaces/http"
	"github.com/stretchr/testify/assert"
)

func setupHostHandler() (*backupHttp.HostHandler, *memory.HostRepositoryMemory) {
	hostRepo := memory.NewHostRepositoryMemory()
	backupRepo := memory.NewBackupRepositoryMemory()
	service := application.NewHostService(hostRepo, backupRepo)
	handler := backupHttp.NewHostHandler(service)
	return handler, hostRepo
}

func TestHostHandler_Create(t *testing.T) {
	handler, _ := setupHostHandler()

	reqBody := dto.CreateHostRequest{
		Name:          "Test Host",
		Hostname:      "test.example.com",
		User:          "testuser",
		Port:          22,
		Path:          "/test/path",
		IsWorkstation: false,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/hosts", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp dto.HostResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, reqBody.Name, resp.Name)
	assert.Equal(t, reqBody.Hostname, resp.Hostname)
	assert.Equal(t, reqBody.User, resp.User)
	assert.Equal(t, reqBody.Port, resp.Port)
	assert.Equal(t, reqBody.Path, resp.Path)
	assert.Equal(t, reqBody.IsWorkstation, resp.IsWorkstation)
}

func TestHostHandler_List(t *testing.T) {
	handler, hostRepo := setupHostHandler()

	// Create a host
	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)
	_ = hostRepo.Save(context.Background(), host)

	req, _ := http.NewRequest("GET", "/hosts", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var hosts []dto.HostResponse
	err := json.Unmarshal(rr.Body.Bytes(), &hosts)
	assert.NoError(t, err)
	// Default repository has 2 hosts, plus we added 1
	assert.GreaterOrEqual(t, len(hosts), 1)
}

func TestHostHandler_Get(t *testing.T) {
	handler, hostRepo := setupHostHandler()

	// Create a host
	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)
	_ = hostRepo.Save(context.Background(), host)

	req, _ := http.NewRequest("GET", "/hosts/"+host.ID().String(), nil)
	req.SetPathValue("id", host.ID().String())
	rr := httptest.NewRecorder()

	handler.Get(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp dto.HostResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, host.ID().String(), resp.ID)
	assert.Equal(t, host.Name(), resp.Name)
}

func TestHostHandler_Update(t *testing.T) {
	handler, hostRepo := setupHostHandler()

	// Create a host
	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)
	_ = hostRepo.Save(context.Background(), host)

	// Update it
	updateBody := dto.UpdateHostRequest{
		ID:            host.ID().String(),
		Name:          "Updated Host",
		Hostname:      "updated.example.com",
		User:          "updateduser",
		Port:          2222,
		Path:          "/updated/path",
		IsWorkstation: true,
	}
	body, _ := json.Marshal(updateBody)

	req, _ := http.NewRequest("PUT", "/hosts/"+host.ID().String(), bytes.NewBuffer(body))
	req.SetPathValue("id", host.ID().String())
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp dto.HostResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Host", resp.Name)
	assert.Equal(t, "updated.example.com", resp.Hostname)
	assert.Equal(t, "updateduser", resp.User)
	assert.Equal(t, 2222, resp.Port)
	assert.Equal(t, "/updated/path", resp.Path)
	assert.True(t, resp.IsWorkstation)
}

func TestHostHandler_Delete(t *testing.T) {
	handler, hostRepo := setupHostHandler()

	// Create a host
	host := entities.NewHost("Test Host", "test.example.com", "user", 22, "path", false)
	_ = hostRepo.Save(context.Background(), host)

	// Delete it
	req, _ := http.NewRequest("DELETE", "/hosts/"+host.ID().String(), nil)
	req.SetPathValue("id", host.ID().String())
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify it's gone
	_, err := hostRepo.Get(context.Background(), host.ID())
	assert.Error(t, err)
}

func TestHostHandler_Create_InvalidJSON(t *testing.T) {
	handler, _ := setupHostHandler()

	req, _ := http.NewRequest("POST", "/hosts", bytes.NewBuffer([]byte("invalid json")))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHostHandler_Get_NotFound(t *testing.T) {
	handler, _ := setupHostHandler()

	req, _ := http.NewRequest("GET", "/hosts/nonexistent-id", nil)
	req.SetPathValue("id", "nonexistent-id")
	rr := httptest.NewRecorder()

	handler.Get(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
