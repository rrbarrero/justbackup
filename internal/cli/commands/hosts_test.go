package commands

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHostsCommandListsHosts(t *testing.T) {
	withTempHome(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hosts" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[
			{"id":"h1","name":"srv","host":"example.com","port":22,"user":"root"}
		]`))
	}))
	defer server.Close()

	writeTestConfig(t, server.URL)

	output := captureOutput(t, HostsCommand)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("unexpected output: %s", output)
	}
	header := lines[0]
	if !strings.Contains(header, "ID") || !strings.Contains(header, "HOST") || !strings.Contains(header, "PORT") {
		t.Fatalf("missing header fields: %s", header)
	}
	row := lines[1]
	if !strings.Contains(row, "h1") || !strings.Contains(row, "example.com") || !strings.Contains(row, "root") {
		t.Fatalf("missing host row values: %s", row)
	}
}
