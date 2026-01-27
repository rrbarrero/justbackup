package commands

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBackupsCommandListsBackups(t *testing.T) {
	withTempHome(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/backups" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[
			{"id":"b1","host_id":"h1","path":"/src","destination":"dest","status":"ok","schedule":"0 0 * * *","last_run":"2024-01-02T03:04:05Z"}
		]`))
	}))
	defer server.Close()

	writeTestConfig(t, server.URL)

	output := captureOutput(t, func() {
		BackupsCommand("")
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("unexpected output: %s", output)
	}
	header := lines[0]
	if !strings.Contains(header, "HOST ID") || !strings.Contains(header, "LAST RUN") {
		t.Fatalf("missing header fields: %s", header)
	}
	row := lines[1]
	if !strings.Contains(row, "b1") || !strings.Contains(row, "h1") || !strings.Contains(row, "/src") {
		t.Fatalf("missing backup row values: %s", row)
	}
	if !strings.Contains(row, "2024-01-02T03:04:05Z") {
		t.Fatalf("missing last run: %s", row)
	}
}

func TestBackupsCommandWithHostFilter(t *testing.T) {
	withTempHome(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/backups" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("host_id"); got != "host-1" {
			t.Fatalf("unexpected host_id: %s", got)
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	writeTestConfig(t, server.URL)

	output := captureOutput(t, func() {
		BackupsCommand("host-1")
	})

	if !strings.Contains(output, "No backups found.") {
		t.Fatalf("unexpected output: %s", output)
	}
}
