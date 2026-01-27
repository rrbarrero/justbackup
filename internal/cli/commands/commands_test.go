package commands

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/cli/config"
)

func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func writeTestConfig(t *testing.T, url string) {
	t.Helper()
	cfg := &config.Config{
		URL:        url,
		Token:      "test-token",
		IgnoreCert: false,
	}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
}

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	return string(data)
}

func withArgs(t *testing.T, args []string, fn func()) {
	t.Helper()
	old := os.Args
	os.Args = args
	defer func() { os.Args = old }()
	fn()
}

func withStdin(t *testing.T, input string, fn func()) {
	t.Helper()
	old := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	_, _ = io.Copy(w, bytes.NewBufferString(input))
	_ = w.Close()
	os.Stdin = r
	defer func() { os.Stdin = old }()
	fn()
}

// Manual Mocks

type MockSSHService struct {
	InstallKeyFunc func(host string, port int, user string, password string, publicKey string) error
}

func (m *MockSSHService) InstallKey(host string, port int, user string, password string, publicKey string) error {
	if m.InstallKeyFunc != nil {
		return m.InstallKeyFunc(host, port, user, password, publicKey)
	}
	return nil
}

type MockAPIService struct {
	GetSSHKeyFunc      func() (string, error)
	RegisterHostFunc   func(req dto.CreateHostRequest) error
	RequestRestoreFunc func(backupID string, req dto.RestoreRequest) (string, error)
}

func (m *MockAPIService) GetSSHKey() (string, error) {
	if m.GetSSHKeyFunc != nil {
		return m.GetSSHKeyFunc()
	}
	return "test-public-key", nil
}

func (m *MockAPIService) RegisterHost(req dto.CreateHostRequest) error {
	if m.RegisterHostFunc != nil {
		return m.RegisterHostFunc(req)
	}
	return nil
}

func (m *MockAPIService) RequestRestore(backupID string, req dto.RestoreRequest) (string, error) {
	if m.RequestRestoreFunc != nil {
		return m.RequestRestoreFunc(backupID, req)
	}
	return "task-123", nil
}

type MockNetService struct {
	GetLocalIPFunc        func() string
	ListenTCPFunc         func() (string, int, io.Closer, error)
	AcceptAndValidateFunc func(listener io.Closer, token string) (io.ReadCloser, error)
	ExtractTarGzFunc      func(r io.Reader, dest string) error
}

func (m *MockNetService) GetLocalIP() string {
	if m.GetLocalIPFunc != nil {
		return m.GetLocalIPFunc()
	}
	return "127.0.0.1"
}

func (m *MockNetService) ListenTCP() (string, int, io.Closer, error) {
	if m.ListenTCPFunc != nil {
		return m.ListenTCPFunc()
	}
	return "127.0.0.1", 8080, &mockCloser{}, nil
}

func (m *MockNetService) AcceptAndValidate(listener io.Closer, token string) (io.ReadCloser, error) {
	if m.AcceptAndValidateFunc != nil {
		return m.AcceptAndValidateFunc(listener, token)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *MockNetService) ExtractTarGz(r io.Reader, dest string) error {
	if m.ExtractTarGzFunc != nil {
		return m.ExtractTarGzFunc(r, dest)
	}
	return nil
}

type mockCloser struct{}

func (m *mockCloser) Close() error { return nil }
