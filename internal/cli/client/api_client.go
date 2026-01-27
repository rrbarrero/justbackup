package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rrbarrero/justbackup/internal/cli/config"
)

type Client struct {
	config *config.Config
	http   *http.Client
}

func NewClient(cfg *config.Config) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.IgnoreCert},
	}
	return &Client{
		config: cfg,
		http: &http.Client{
			Timeout:   35 * time.Second,
			Transport: tr,
		},
	}
}

func (c *Client) Get(path string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.config.URL, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	return body, nil
}

func (c *Client) Post(path string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.config.URL, path)
	fmt.Printf("Debug: POST %s\n", url)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(resBody))
	}

	return resBody, nil
}
