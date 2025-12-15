package service

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/forfire912/machineServer/internal/adapter"
	"github.com/forfire912/machineServer/internal/config"
	"github.com/forfire912/machineServer/internal/model"
)

// Ensure MockAdapter implements adapter.Adapter
var _ adapter.Adapter = (*MockAdapter)(nil)

// MockAdapter implements adapter.Adapter for testing
type MockAdapter struct {
	Calls map[string]int
}

func NewMockAdapter() *MockAdapter {
	return &MockAdapter{
		Calls: make(map[string]int),
	}
}

func (m *MockAdapter) GetCapabilities() (*model.Capability, error) {
	return &model.Capability{Backend: "mock"}, nil
}

func (m *MockAdapter) StartSession(ctx context.Context, session *model.Session, config *model.BoardConfig, consoleOut io.Writer) error {
	m.Calls["StartSession"]++
	session.GDBPort = 1234
	session.MonitorPort = 5678
	return nil
}

func (m *MockAdapter) StopSession(ctx context.Context, sessionID string) error {
	m.Calls["StopSession"]++
	return nil
}

func (m *MockAdapter) ResetSession(ctx context.Context, sessionID string) error {
	m.Calls["ResetSession"]++
	return nil
}

func (m *MockAdapter) LoadProgram(ctx context.Context, sessionID string, programPath string) error {
	m.Calls["LoadProgram"]++
	return nil
}

func (m *MockAdapter) ExecuteProgram(ctx context.Context, sessionID string) error {
	m.Calls["ExecuteProgram"]++
	return nil
}

func (m *MockAdapter) PauseExecution(ctx context.Context, sessionID string) error {
	m.Calls["PauseExecution"]++
	return nil
}

func (m *MockAdapter) ResumeExecution(ctx context.Context, sessionID string) error {
	m.Calls["ResumeExecution"]++
	return nil
}

func (m *MockAdapter) GetGDBPort(sessionID string) (int, error) {
	m.Calls["GetGDBPort"]++
	return 1234, nil
}

func (m *MockAdapter) CreateSnapshot(ctx context.Context, sessionID string, snapshotPath string) error {
	m.Calls["CreateSnapshot"]++
	return nil
}

func (m *MockAdapter) RestoreSnapshot(ctx context.Context, sessionID string, snapshotPath string) error {
	m.Calls["RestoreSnapshot"]++
	return nil
}

func (m *MockAdapter) GetConsoleOutput(ctx context.Context, sessionID string) (string, error) {
	m.Calls["GetConsoleOutput"]++
	return "mock output", nil
}

func (m *MockAdapter) Step(ctx context.Context, sessionID string, steps int) error {
	m.Calls["Step"]++
	return nil
}

func (m *MockAdapter) StartCoverage(ctx context.Context, sessionID string, outputPath string) error {
	m.Calls["StartCoverage"]++
	return nil
}

func (m *MockAdapter) StopCoverage(ctx context.Context, sessionID string) error {
	m.Calls["StopCoverage"]++
	return nil
}

func (m *MockAdapter) RunForTime(ctx context.Context, sessionID string, duration time.Duration) error {
	m.Calls["RunForTime"]++
	return nil
}

func (m *MockAdapter) InjectEvent(ctx context.Context, sessionID string, eventType string, data map[string]interface{}) error {
	m.Calls["InjectEvent"]++
	return nil
}

func TestFullCapabilities(t *testing.T) {
	// Setup Service with Mock Adapter
	cfg := &config.Config{
		Database: config.DatabaseConfig{Driver: "sqlite", DSN: ":memory:"},
		Storage:  config.StorageConfig{BasePath: "/tmp/test-caps"},
		Resources: config.ResourcesConfig{MaxSessions: 10},
	}
	
	svc, err := NewService(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	mockAdp := NewMockAdapter()
	svc.adapters[model.BackendQEMU] = mockAdp
	svc.adapters[model.BackendRenode] = mockAdp
	svc.adapters[model.BackendOpenOCD] = mockAdp

	ctx := context.Background()

	// ==========================================
	// 1. Session Management & Board Config
	// ==========================================
	t.Log("Testing Session Management...")
	boardCfg := &model.BoardConfig{Board: "test-board"}
	session, err := svc.CreateSession(ctx, "test-session", model.BackendQEMU, boardCfg)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if mockAdp.Calls["StartSession"] != 1 {
		t.Error("StartSession not called on adapter")
	}

	// Verify Board Config stored
	var storedCfg model.BoardConfig
	json.Unmarshal([]byte(session.BoardConfig), &storedCfg)
	if storedCfg.Board != "test-board" {
		t.Errorf("Board config not stored correctly")
	}

	// ==========================================
	// 2. Simulation Control
	// ==========================================
	t.Log("Testing Simulation Control...")
	if err := svc.PowerOn(ctx, session.ID); err != nil {
		t.Errorf("PowerOn failed: %v", err)
	}
	if mockAdp.Calls["ExecuteProgram"] != 1 {
		t.Error("ExecuteProgram not called")
	}

	if err := svc.PowerOff(ctx, session.ID); err != nil {
		t.Errorf("PowerOff failed: %v", err)
	}
	if mockAdp.Calls["PauseExecution"] != 1 {
		t.Error("PauseExecution not called")
	}

	if err := svc.Reset(ctx, session.ID); err != nil {
		t.Errorf("Reset failed: %v", err)
	}
	if mockAdp.Calls["ResetSession"] != 1 {
		t.Error("ResetSession not called")
	}

	// ==========================================
	// 3. Coverage
	// ==========================================
	t.Log("Testing Coverage...")
	if err := svc.StartCoverage(ctx, session.ID); err != nil {
		t.Errorf("StartCoverage failed: %v", err)
	}
	if mockAdp.Calls["StartCoverage"] != 1 {
		t.Error("StartCoverage not called")
	}

	if err := svc.StopCoverage(ctx, session.ID); err != nil {
		t.Errorf("StopCoverage failed: %v", err)
	}
	if mockAdp.Calls["StopCoverage"] != 1 {
		t.Error("StopCoverage not called")
	}

	// ==========================================
	// 4. Co-Simulation (Multi-node)
	// ==========================================
	t.Log("Testing Co-Simulation...")
	
	// Create another session for co-sim
	sess2, _ := svc.CreateSession(ctx, "node-2", model.BackendRenode, nil)

	comps := []model.CoSimComponent{
		{Type: "qemu", SessionID: session.ID},
		{Type: "renode", SessionID: sess2.ID},
	}
	
	coSession, err := svc.CreateCoSimSession(ctx, comps)
	if err != nil {
		t.Fatalf("CreateCoSimSession failed: %v", err)
	}

	// Start CoSim
	svc.StartCoSimSession(ctx, coSession.ID)

	// Scheme 2: SyncStep
	if err := svc.SyncStep(ctx, coSession.ID, 100); err != nil {
		t.Errorf("SyncStep failed: %v", err)
	}
	if mockAdp.Calls["Step"] != 2 { // Called for both components
		t.Errorf("Step called %d times, expected 2", mockAdp.Calls["Step"])
	}

	// Scheme 3: SyncTime
	if err := svc.SyncTime(ctx, coSession.ID, 1*time.Millisecond); err != nil {
		t.Errorf("SyncTime failed: %v", err)
	}
	if mockAdp.Calls["RunForTime"] != 2 {
		t.Errorf("RunForTime called %d times, expected 2", mockAdp.Calls["RunForTime"])
	}

	// Scheme 4: InjectEvent
	// Find component ID for session 1
	var compID string
	for _, c := range coSession.Components {
		if c.SessionID == session.ID {
			compID = c.ID
			break
		}
	}
	
	err = svc.InjectCoSimEvent(ctx, coSession.ID, compID, "gpio", map[string]interface{}{"state": true})
	if err != nil {
		t.Errorf("InjectCoSimEvent failed: %v", err)
	}
	if mockAdp.Calls["InjectEvent"] != 1 {
		t.Error("InjectEvent not called")
	}

	// ==========================================
	// 5. Cleanup
	// ==========================================
	t.Log("Testing Cleanup...")
	if err := svc.DeleteCoSimSession(ctx, coSession.ID); err != nil {
		t.Errorf("DeleteCoSimSession failed: %v", err)
	}
	
	if err := svc.DeleteSession(ctx, session.ID); err != nil {
		t.Errorf("DeleteSession failed: %v", err)
	}
	if mockAdp.Calls["StopSession"] < 1 {
		t.Error("StopSession not called during delete")
	}
}
