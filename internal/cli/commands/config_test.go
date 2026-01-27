package commands

import (
	"strings"
	"testing"

	"github.com/rrbarrero/justbackup/internal/cli/config"
)

func TestConfigCommandSavesConfig(t *testing.T) {
	withTempHome(t)

	input := "\nmytoken\ny\n"
	output := captureOutput(t, func() {
		withStdin(t, input, func() {
			ConfigCommand("example.com")
		})
	})

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.URL != "https://example.com/api/v1" {
		t.Fatalf("unexpected url: %s", cfg.URL)
	}
	if cfg.Token != "mytoken" {
		t.Fatalf("unexpected token: %s", cfg.Token)
	}
	if !cfg.IgnoreCert {
		t.Fatalf("expected ignore cert true")
	}
	if !strings.Contains(output, "Configuration saved successfully") {
		t.Fatalf("unexpected output: %s", output)
	}
}
