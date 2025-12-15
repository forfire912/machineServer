package service

import (
	"context"
	"os"
	"testing"

	"github.com/forfire912/machineServer/internal/config"
	"github.com/forfire912/machineServer/internal/model"
)

func TestCoSimulation(t *testing.T) {
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

	// 1. Create CoSim Session
	components := []model.CoSimComponent{
		{
			Type:   "qemu",
			Config: "{}",
		},
		{
			Type:   "renode",
			Config: "{}",
		},
	}

	session, err := svc.CreateCoSimSession(context.Background(), components)
	if err != nil {
		t.Fatalf("Failed to create cosim session: %v", err)
	}

	if session.ID == "" {
		t.Fatal("Session ID is empty")
	}
	if len(session.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(session.Components))
	}
	if session.Status != "created" {
		t.Errorf("Expected status created, got %s", session.Status)
	}

	// 2. Start Session
	err = svc.StartCoSimSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("Failed to start cosim session: %v", err)
	}

	// Verify status
	session, err = svc.GetCoSimSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get cosim session: %v", err)
	}
	if session.Status != "running" {
		t.Errorf("Expected status running, got %s", session.Status)
	}

	// 3. Sync Step
	// Note: This might fail if backends are not actually running, but the service logic should hold
	// The service tries to find sessions for components, but we haven't linked them to real sessions yet.
	// In CreateCoSimSession, we just created components records.
	// The SyncStep logic checks: if comp.SessionID != "" -> find session -> step.
	// Since SessionID is empty, it should just skip and update the CoSimSession counters.
	
	err = svc.SyncStep(context.Background(), session.ID, 100)
	if err != nil {
		t.Fatalf("Failed to sync step: %v", err)
	}

	session, err = svc.GetCoSimSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get cosim session: %v", err)
	}
	if session.SyncCount != 100 {
		t.Errorf("Expected sync count 100, got %d", session.SyncCount)
	}

	// 4. Stop Session
	err = svc.StopCoSimSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("Failed to stop cosim session: %v", err)
	}

	session, err = svc.GetCoSimSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get cosim session: %v", err)
	}
	if session.Status != "stopped" {
		t.Errorf("Expected status stopped, got %s", session.Status)
	}

	// 5. List Sessions
	sessions, total, err := svc.ListCoSimSessions(1, 10)
	if err != nil {
		t.Fatalf("Failed to list cosim sessions: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected 1 session, got %d", total)
	}
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session in list, got %d", len(sessions))
	}

	// 6. Delete Session
	err = svc.DeleteCoSimSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("Failed to delete cosim session: %v", err)
	}

	// Verify deletion
	_, err = svc.GetCoSimSession(session.ID)
	if err == nil {
		t.Error("Expected error getting deleted session, got nil")
	}
}
