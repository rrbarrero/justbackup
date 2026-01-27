package commands

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunBackupCommand(t *testing.T) {
	withTempHome(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/backups/b1/run" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		_, _ = w.Write([]byte(`{"task_id":"t1"}`))
	}))
	defer server.Close()

	writeTestConfig(t, server.URL)

	output := captureOutput(t, func() {
		RunBackupCommand("b1")
	})

	if !strings.Contains(output, "Backup triggered successfully. Task ID: t1") {
		t.Fatalf("unexpected output: %s", output)
	}
}
