package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/container"
)

func setupTestServer() http.Handler {
	// Set required environment variables for testing (use dev to get in-memory repos)
	_ = os.Setenv("ENVIRONMENT", "dev")
	_ = os.Setenv("JWT_SECRET", "test-secret-key")
	_ = os.Setenv("REDIS_HOST", "localhost")
	_ = os.Setenv("REDIS_PORT", "6379")
	_ = os.Setenv("ENCRYPTION_KEY", "test-encryption-key")

	// Initialize container to get the handler
	container, err := container.InitializeContainer()
	if err != nil {
		panic(err) // In tests, we can panic if initialization fails
	}

	return container.GetHandler()
}

func TestHealthCheck(t *testing.T) {
	handler := setupTestServer()
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("could not send GET request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}

	if string(body) != "OK" {
		t.Errorf("expected body 'OK'; got '%s'", string(body))
	}
}

func TestProtectedRoutes(t *testing.T) {
	handler := setupTestServer()
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("it should not allow access to /api/v1/backups without token", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/backups")
		if err != nil {
			t.Fatalf("could not send GET request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status Unauthorized; got %v", resp.Status)
		}
	})

	t.Run("it should not allow access to /api/v1/hosts without token", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/hosts")
		if err != nil {
			t.Fatalf("could not send GET request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status Unauthorized; got %v", resp.Status)
		}
	})
}

func TestPublicRoutes(t *testing.T) {
	handler := setupTestServer()
	server := httptest.NewServer(handler)
	defer server.Close()

	// 1. Setup a user first
	setupBody := `{"username": "testuser", "password": "password"}`
	resp, err := http.Post(server.URL+"/api/v1/setup", "application/json", strings.NewReader(setupBody))
	if err != nil {
		t.Fatalf("could not send POST request to /setup: %v", err)
	}
	// It could be that the user is already created if tests run multiple times, so we accept 201 or 400.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status Created or Bad Request from /setup; got %v", resp.Status)
	}
	_ = resp.Body.Close()

	t.Run("it should allow login with correct credentials", func(t *testing.T) {
		loginBody := `{"username": "testuser", "password": "password"}`
		resp, err := http.Post(server.URL+"/api/v1/login", "application/json", strings.NewReader(loginBody))
		if err != nil {
			t.Fatalf("could not send POST request to /login: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("expected status OK for login; got %v. Body: %s", resp.Status, body)
		}
	})

	t.Run("it should not allow login with incorrect credentials", func(t *testing.T) {
		loginBody := `{"username": "testuser", "password": "wrongpassword"}`
		resp, err := http.Post(server.URL+"/api/v1/login", "application/json", strings.NewReader(loginBody))
		if err != nil {
			t.Fatalf("could not send POST request to /login: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status Unauthorized for login; got %v", resp.Status)
		}
	})

	t.Run("it should still fail for empty body", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/api/v1/login", "application/json", strings.NewReader(`{}`))
		if err != nil {
			t.Fatalf("could not send POST request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected endpoint to return unauthorized for empty body, got %v", resp.Status)
		}
	})
}
