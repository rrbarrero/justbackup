package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rrbarrero/justbackup/internal/cli/config"
)

func ConfigCommand(defaultURL string) {
	reader := bufio.NewReader(os.Stdin)

	// Normalize default URL for display
	displayURL := defaultURL
	if displayURL != "" {
		if !strings.HasPrefix(displayURL, "http://") && !strings.HasPrefix(displayURL, "https://") {
			displayURL = "https://" + displayURL
		}
		if !strings.Contains(displayURL, "/api/v1") {
			displayURL = strings.TrimSuffix(displayURL, "/") + "/api/v1"
		}
	}

	// URL Prompt
	fmt.Printf("Backend API URL [%s]: ", displayURL)
	urlInput, _ := reader.ReadString('\n')
	urlInput = strings.TrimSpace(urlInput)
	if urlInput == "" {
		urlInput = displayURL
	}

	// Ensure protocol and /api/v1 suffix on user input too
	if urlInput != "" {
		if !strings.HasPrefix(urlInput, "http://") && !strings.HasPrefix(urlInput, "https://") {
			urlInput = "https://" + urlInput
		}
		if !strings.Contains(urlInput, "/api/v1") {
			urlInput = strings.TrimSuffix(urlInput, "/") + "/api/v1"
		}
	}

	// Token Prompt
	fmt.Print("API Token: ")
	tokenInput, _ := reader.ReadString('\n')
	tokenInput = strings.TrimSpace(tokenInput)

	if tokenInput == "" {
		fmt.Println("Error: API Token is required.")
		return
	}

	// Ignore Cert Prompt
	fmt.Print("Ignore SSL Certificate Errors? [y/N]: ")
	ignoreCertInput, _ := reader.ReadString('\n')
	ignoreCertInput = strings.TrimSpace(strings.ToLower(ignoreCertInput))
	ignoreCert := ignoreCertInput == "y" || ignoreCertInput == "yes"

	// Save Config
	cfg := &config.Config{
		URL:        urlInput,
		Token:      tokenInput,
		IgnoreCert: ignoreCert,
	}

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return
	}

	fmt.Printf("Configuration saved successfully. Using API: %s\n", urlInput)
}
