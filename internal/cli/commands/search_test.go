package commands

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchCommand(t *testing.T) {
	withTempHome(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/files/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("pattern") != "foo bar" {
			t.Fatalf("unexpected pattern: %s", r.URL.Query().Get("pattern"))
		}
		_, _ = w.Write([]byte(`[
			{"path":"/etc/hosts","backup":{"id":"b1","host_name":"srv","destination":"dest","status":"ok"}}
		]`))
	}))
	defer server.Close()

	writeTestConfig(t, server.URL)

	output := captureOutput(t, func() {
		SearchCommand("foo bar")
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("unexpected output: %s", output)
	}
	header := lines[0]
	if !strings.Contains(header, "FILE PATH") || !strings.Contains(header, "DESTINATION") {
		t.Fatalf("missing header fields: %s", header)
	}
	row := lines[1]
	if !strings.Contains(row, "/etc/hosts") || !strings.Contains(row, "srv") || !strings.Contains(row, "b1") {
		t.Fatalf("missing result row: %s", row)
	}
}
