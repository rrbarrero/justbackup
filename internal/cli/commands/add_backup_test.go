package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
)

func TestAddBackupCommandPostsRequest(t *testing.T) {
	withTempHome(t)

	var gotReq dto.CreateBackupRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/backups" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Fatalf("unexpected auth header: %s", auth)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(body, &gotReq); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	writeTestConfig(t, server.URL)

	args := []string{"justbackup", "add-backup", "--host-id", "host-1", "--path", "/src", "--dest", "daily", "--excludes", "tmp,cache"}
	output := captureOutput(t, func() {
		withArgs(t, args, AddBackupCommand)
	})

	if !strings.Contains(output, "Backup task created successfully") {
		t.Fatalf("unexpected output: %s", output)
	}

	if gotReq.HostID != "host-1" || gotReq.Path != "/src" || gotReq.Destination != "daily" {
		t.Fatalf("unexpected request: %+v", gotReq)
	}
	if gotReq.Schedule != "0 0 * * *" {
		t.Fatalf("unexpected schedule: %s", gotReq.Schedule)
	}
	if !gotReq.Incremental {
		t.Fatalf("expected incremental true")
	}
	if len(gotReq.Excludes) != 2 || gotReq.Excludes[0] != "tmp" || gotReq.Excludes[1] != "cache" {
		t.Fatalf("unexpected excludes: %+v", gotReq.Excludes)
	}
}
