package config

import (
	"fmt"
	"os"
	"strings"
)

// ServerConfig holds configuration for the server
type ServerConfig struct {
	Environment       string
	JWTSecret         string
	RedisHost         string
	RedisPort         string
	CORSAllowedOrigin string
	EncryptionKey     string
	ServerPort        string // Added for better server configuration
}

// WorkerConfig holds configuration for the worker
type WorkerConfig struct {
	Environment         string
	RedisURL            string
	SSHKeyPath          string
	HostBackupRoot      string
	ContainerBackupRoot string
	EncryptionKey       string
	BackendURL          string
}

// ConfigService provides methods to access configuration
type ConfigService struct {
	serverConfig *ServerConfig
	workerConfig *WorkerConfig
}

// NewConfigService creates a new configuration service
func NewConfigService() *ConfigService {
	return &ConfigService{}
}

// LoadServerConfig loads and validates server configuration
func (cs *ConfigService) LoadServerConfig() (*ServerConfig, error) {
	if cs.serverConfig != nil {
		return cs.serverConfig, nil // Return cached config if already loaded
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "production"
	}

	config := &ServerConfig{
		Environment:       env,
		JWTSecret:         os.Getenv("JWT_SECRET"),
		RedisHost:         os.Getenv("REDIS_HOST"),
		RedisPort:         os.Getenv("REDIS_PORT"),
		CORSAllowedOrigin: os.Getenv("CORS_ALLOWED_ORIGIN"),
		EncryptionKey:     os.Getenv("ENCRYPTION_KEY"),
		ServerPort:        getEnv("SERVER_PORT", "8080"), // Default to 8080
	}

	// Only validate in production mode
	if env != "dev" && env != "development" {
		var missing []string

		if config.JWTSecret == "" {
			missing = append(missing, "JWT_SECRET")
		}
		if config.RedisHost == "" {
			missing = append(missing, "REDIS_HOST")
		}
		if config.RedisPort == "" {
			missing = append(missing, "REDIS_PORT")
		}
		if config.EncryptionKey == "" {
			missing = append(missing, "ENCRYPTION_KEY")
		}

		if len(missing) > 0 {
			return nil, fmt.Errorf(`
┌─────────────────────────────────────────────────────────────────┐
│ ❌ CRITICAL CONFIGURATION ERROR                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ The following REQUIRED environment variables are NOT SET:      │
│                                                                 │
│   • %s
│                                                                 │
│ The server CANNOT START without these variables.               │
│                                                                 │
│ Please set them in your environment or docker-compose.yml      │
│ and restart the server.                                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
`, strings.Join(missing, "\n│   • "))
		}
	}

	// Set defaults for optional variables
	if config.CORSAllowedOrigin == "" {
		config.CORSAllowedOrigin = "http://localhost:3000"
	}

	cs.serverConfig = config
	return config, nil
}

// LoadWorkerConfig loads and validates worker configuration
func (cs *ConfigService) LoadWorkerConfig() (*WorkerConfig, error) {
	if cs.workerConfig != nil {
		return cs.workerConfig, nil // Return cached config if already loaded
	}

	CONTAINER_BACKUP_ROOT := "/mnt/backups"

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "production"
	}

	config := &WorkerConfig{
		Environment:         env,
		RedisURL:            os.Getenv("REDIS_URL"),
		SSHKeyPath:          os.Getenv("SSH_KEY_PATH"),
		HostBackupRoot:      os.Getenv("BACKUP_ROOT"),
		ContainerBackupRoot: CONTAINER_BACKUP_ROOT,
		EncryptionKey:       os.Getenv("ENCRYPTION_KEY"),
		BackendURL:          os.Getenv("BACKEND_INTERNAL_URL"),
	}

	// Only validate in production mode
	if env != "dev" && env != "development" {
		var missing []string

		if config.RedisURL == "" {
			missing = append(missing, "REDIS_URL")
		}
		if config.SSHKeyPath == "" {
			missing = append(missing, "SSH_KEY_PATH")
		}
		if config.HostBackupRoot == "" {
			missing = append(missing, "BACKUP_ROOT")
		}

		if len(missing) > 0 {
			return nil, fmt.Errorf(`
┌─────────────────────────────────────────────────────────────────┐
│ ❌ CRITICAL CONFIGURATION ERROR                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ The following REQUIRED environment variables are NOT SET:      │
│                                                                 │
│   • %s
│                                                                 │
│ The worker CANNOT START without these variables.               │
│                                                                 │
│ Please set them in your environment or docker-compose.yml      │
│ and restart the worker.                                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
`, strings.Join(missing, "\n│   • "))
		}
	}

	// Set defaults for dev mode
	if env == "dev" || env == "development" {
		if config.RedisURL == "" {
			config.RedisURL = "redis:6379"
		}
		if config.SSHKeyPath == "" {
			config.SSHKeyPath = "/home/backup/.ssh/id_ed25519"
		}
		if config.HostBackupRoot == "" {
			config.HostBackupRoot = "/mnt/backups"
		}
		if config.BackendURL == "" {
			config.BackendURL = "http://server:8080"
		}
	}

	cs.workerConfig = config
	return config, nil
}

// GetServerConfig returns the loaded server configuration
func (cs *ConfigService) GetServerConfig() *ServerConfig {
	return cs.serverConfig
}

// GetWorkerConfig returns the loaded worker configuration
func (cs *ConfigService) GetWorkerConfig() *WorkerConfig {
	return cs.workerConfig
}

// LoadServerConfig is a convenience function that creates and uses a ConfigService
func LoadServerConfig() (*ServerConfig, error) {
	service := NewConfigService()
	return service.LoadServerConfig()
}

// LoadWorkerConfig is a convenience function that creates and uses a ConfigService
func LoadWorkerConfig() (*WorkerConfig, error) {
	service := NewConfigService()
	return service.LoadWorkerConfig()
}

// Helper function to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
