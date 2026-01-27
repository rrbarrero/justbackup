package module

import (
	"context"
	"fmt"
	"log"

	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
)

// StartupModule defines the interface for a startup module
type StartupModule interface {
	Name() string
	Initialize(cfg *config.ServerConfig) error
	Start(ctx context.Context) error
	Stop() error
}

// ModuleManager manages the startup and shutdown of modules
type ModuleManager struct {
	modules []StartupModule
}

// NewModuleManager creates a new module manager
func NewModuleManager() *ModuleManager {
	return &ModuleManager{
		modules: make([]StartupModule, 0),
	}
}

// RegisterModule registers a new startup module
func (mm *ModuleManager) RegisterModule(module StartupModule) {
	mm.modules = append(mm.modules, module)
	log.Printf("Registered module: %s", module.Name())
}

// InitializeAll initializes all registered modules
func (mm *ModuleManager) InitializeAll(cfg *config.ServerConfig) error {
	for _, module := range mm.modules {
		log.Printf("Initializing module: %s", module.Name())
		if err := module.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize module %s: %w", module.Name(), err)
		}
	}
	return nil
}

// StartAll starts all registered modules
func (mm *ModuleManager) StartAll(ctx context.Context) error {
	for _, module := range mm.modules {
		log.Printf("Starting module: %s", module.Name())
		if err := module.Start(ctx); err != nil {
			return fmt.Errorf("failed to start module %s: %w", module.Name(), err)
		}
	}
	return nil
}

// StopAll stops all registered modules
func (mm *ModuleManager) StopAll() error {
	for _, module := range mm.modules {
		log.Printf("Stopping module: %s", module.Name())
		if err := module.Stop(); err != nil {
			log.Printf("Error stopping module %s: %v", module.Name(), err)
		}
	}
	return nil
}

// ModuleNames returns the names of all registered modules
func (mm *ModuleManager) ModuleNames() []string {
	names := make([]string, len(mm.modules))
	for i, module := range mm.modules {
		names[i] = module.Name()
	}
	return names
}
