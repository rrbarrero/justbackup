package commands

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestNetService_GetLocalIP(t *testing.T) {
	svc := NewNetService()
	ip := svc.GetLocalIP()
	if ip == "" {
		t.Error("expected non-empty IP")
	}
}

func TestNetService_ListenTCP(t *testing.T) {
	svc := NewNetService()
	host, port, listener, err := svc.ListenTCP()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if listener == nil {
		t.Fatal("expected listener to be non-nil")
	}
	defer func() { _ = listener.Close() }()

	if host == "" {
		t.Error("expected non-empty host")
	}
	if port <= 0 {
		t.Errorf("expected positive port, got %d", port)
	}
}

func TestNetService_ExtractTarGz(t *testing.T) {
	svc := NewNetService()
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	tr := tar.NewWriter(gz)

	content := []byte("hello world")
	hdr := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: int64(len(content)),
	}
	if err := tr.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tr.Write(content); err != nil {
		t.Fatal(err)
	}
	_ = tr.Close()
	_ = gz.Close()

	dest := t.TempDir()
	if err := svc.ExtractTarGz(buf, dest); err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Errorf("expected %s, got %s", content, got)
	}
}
