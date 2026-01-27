package http_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	backupHttp "github.com/rrbarrero/justbackup/internal/backup/interfaces/http"
	"github.com/stretchr/testify/assert"
)

func TestSettingsHandler_GetSSHKey_Success(t *testing.T) {
	// Create a temporary SSH key file
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key.pub")
	keyContent := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITest testkey@example.com"
	err := os.WriteFile(keyPath, []byte(keyContent), 0644)
	assert.NoError(t, err)

	// Set environment variable
	t.Setenv("SSH_PUBLIC_KEY_PATH", keyPath)

	handler := backupHttp.NewSettingsHandler()

	req, _ := http.NewRequest("GET", "/settings/ssh-key", nil)
	rr := httptest.NewRecorder()

	handler.GetSSHKey(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), keyContent)
}

func TestSettingsHandler_GetSSHKey_NoPathConfigured(t *testing.T) {
	// Unset environment variable
	t.Setenv("SSH_PUBLIC_KEY_PATH", "")

	handler := backupHttp.NewSettingsHandler()

	req, _ := http.NewRequest("GET", "/settings/ssh-key", nil)
	rr := httptest.NewRecorder()

	handler.GetSSHKey(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "not configured")
}

func TestSettingsHandler_GetSSHKey_FileNotFound(t *testing.T) {
	// Set environment variable to non-existent path
	t.Setenv("SSH_PUBLIC_KEY_PATH", "/nonexistent/path/key.pub")

	handler := backupHttp.NewSettingsHandler()

	req, _ := http.NewRequest("GET", "/settings/ssh-key", nil)
	rr := httptest.NewRecorder()

	handler.GetSSHKey(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to read")
}
