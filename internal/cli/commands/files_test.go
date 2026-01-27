package commands

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFilesCommandListsFiles(t *testing.T) {
	withTempHome(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/backups/b1/files" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[
			{"name":"dir","is_dir":true,"size":0},
			{"name":"file.txt","is_dir":false,"size":1234}
		]`))
	}))
	defer server.Close()

	writeTestConfig(t, server.URL)

	args := []string{"justbackup", "files", "b1"}
	output := captureOutput(t, func() {
		withArgs(t, args, FilesCommand)
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Fatalf("unexpected output: %s", output)
	}
	header := lines[0]
	if !strings.Contains(header, "NAME") || !strings.Contains(header, "TYPE") || !strings.Contains(header, "SIZE") {
		t.Fatalf("missing header fields: %s", header)
	}
	if !strings.Contains(lines[1], "dir") || !strings.Contains(lines[1], "DIR") {
		t.Fatalf("missing dir row: %s", lines[1])
	}
	if !strings.Contains(lines[2], "file.txt") || !strings.Contains(lines[2], "1.2 KB") {
		t.Fatalf("missing file row: %s", lines[2])
	}
}

func TestFormatSize(t *testing.T) {
	cases := map[int64]string{
		0:       "0 B",
		1023:    "1023 B",
		1024:    "1.0 KB",
		1048576: "1.0 MB",
	}
	for size, expect := range cases {
		if got := formatSize(size); got != expect {
			t.Fatalf("size %d: expected %s, got %s", size, expect, got)
		}
	}
}
