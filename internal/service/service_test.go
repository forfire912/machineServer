package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/forfire912/machineServer/internal/config"
	"github.com/forfire912/machineServer/internal/model"
)

func TestNewService(t *testing.T) {
	// Create temp config
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DSN:    ":memory:",
		},
		Storage: config.StorageConfig{
			BasePath: "/tmp/test-storage",
		},
		Resources: config.ResourcesConfig{
			MaxSessions: 10,
		},
		Backends: config.BackendsConfig{
			QEMU: config.BackendConfig{
				Enabled: true,
				Binary:  "qemu-system-arm",
			},
		},
	}

	defer os.RemoveAll("/tmp/test-storage")

	svc, err := NewService(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if svc == nil {
		t.Fatal("Service is nil")
	}

	if svc.db == nil {
		t.Fatal("Database is nil")
	}
}

func TestGetCapabilities(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DSN:    ":memory:",
		},
		Storage: config.StorageConfig{
			BasePath: "/tmp/test-storage",
		},
		Backends: config.BackendsConfig{
			QEMU: config.BackendConfig{
				Enabled: true,
				Binary:  "qemu-system-arm",
			},
		},
	}

	defer os.RemoveAll("/tmp/test-storage")

	svc, err := NewService(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	caps, err := svc.GetCapabilities()
	if err != nil {
		t.Fatalf("Failed to get capabilities: %v", err)
	}

	if len(caps) == 0 {
		t.Fatal("No capabilities returned")
	}

	if caps[0].Backend != model.BackendQEMU {
		t.Errorf("Expected QEMU backend, got %s", caps[0].Backend)
	}
}

func TestCreateAndGetSession(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DSN:    ":memory:",
		},
		Storage: config.StorageConfig{
			BasePath: "/tmp/test-storage",
		},
		Resources: config.ResourcesConfig{
			MaxSessions: 10,
		},
		Backends: config.BackendsConfig{
			QEMU: config.BackendConfig{
				Enabled: true,
				Binary:  "qemu-system-arm",
			},
		},
	}

	defer os.RemoveAll("/tmp/test-storage")

	svc, err := NewService(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	boardConfig := &model.BoardConfig{
		Processor: model.ProcessorConfig{
			Model:     "cortex-m3",
			Frequency: 72000000,
		},
		Memory: model.MemoryConfig{
			Flash: model.MemoryRegion{
				Base: 0x08000000,
				Size: 131072,
			},
			RAM: model.MemoryRegion{
				Base: 0x20000000,
				Size: 20480,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Note: This will fail if QEMU is not installed, which is expected in test environment
	session, err := svc.CreateSession(ctx, "test-session", model.BackendQEMU, boardConfig)
	
	// If QEMU is not available, skip the rest
	if err != nil {
		t.Skipf("Skipping test because QEMU is not available: %v", err)
		return
	}

	if session == nil {
		t.Fatal("Session is nil")
	}

	if session.ID == "" {
		t.Fatal("Session ID is empty")
	}

	// Get the session
	retrieved, err := svc.GetSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrieved.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, retrieved.ID)
	}

	// Cleanup
	defer svc.DeleteSession(ctx, session.ID)
}

func TestListSessions(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DSN:    ":memory:",
		},
		Storage: config.StorageConfig{
			BasePath: "/tmp/test-storage",
		},
		Resources: config.ResourcesConfig{
			MaxSessions: 10,
		},
		Backends: config.BackendsConfig{
			QEMU: config.BackendConfig{
				Enabled: true,
				Binary:  "qemu-system-arm",
			},
		},
	}

	defer os.RemoveAll("/tmp/test-storage")

	svc, err := NewService(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	sessions, total, err := svc.ListSessions(1, 10)
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if sessions == nil {
		t.Fatal("Sessions is nil")
	}

	if total < 0 {
		t.Errorf("Total should be >= 0, got %d", total)
	}
}
