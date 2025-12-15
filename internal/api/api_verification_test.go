package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/forfire912/machineServer/internal/adapter"
	"github.com/forfire912/machineServer/internal/config"
	"github.com/forfire912/machineServer/internal/model"
	"github.com/forfire912/machineServer/internal/service"
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

func TestAPICapabilities(t *testing.T) {
	// Setup Service with Mock Adapter
	cfg := &config.Config{
		Database: config.DatabaseConfig{Driver: "sqlite", DSN: ":memory:"},
		Storage:  config.StorageConfig{BasePath: "/tmp/test-api-caps"},
		Resources: config.ResourcesConfig{MaxSessions: 10},
		Server: config.ServerConfig{Mode: "test"},
	}
	
	svc, err := service.NewService(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	mockAdp := NewMockAdapter()
	svc.SetAdapter(model.BackendQEMU, mockAdp)
	svc.SetAdapter(model.BackendRenode, mockAdp)

	// Setup Router
	router := SetupRouter(cfg, svc, nil)

	// 1. Create Session
	t.Log("Testing API: Create Session")
	reqBody := map[string]interface{}{
		"name": "api-test-session",
		"backend": "qemu",
		"board_config": map[string]string{"board": "test-board"},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}
	
	var sessionResp model.Session
	json.Unmarshal(w.Body.Bytes(), &sessionResp)
	sessionID := sessionResp.ID

	// 2. Power On
	t.Log("Testing API: Power On")
	req, _ = http.NewRequest("POST", "/api/v1/sessions/"+sessionID+"/poweron", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
	if mockAdp.Calls["ExecuteProgram"] != 1 {
		t.Error("ExecuteProgram not called via API")
	}

	// 3. Start Coverage
	t.Log("Testing API: Start Coverage")
	req, _ = http.NewRequest("POST", "/api/v1/sessions/"+sessionID+"/coverage/start", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
	if mockAdp.Calls["StartCoverage"] != 1 {
		t.Error("StartCoverage not called via API")
	}

	// 4. Co-Simulation Create
	t.Log("Testing API: Create Co-Simulation")
	// Create another session
	reqBody2 := map[string]interface{}{
		"name": "api-test-node-2",
		"backend": "renode",
	}
	body2, _ := json.Marshal(reqBody2)
	req2, _ := http.NewRequest("POST", "/api/v1/sessions", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	var sessionResp2 model.Session
	json.Unmarshal(w2.Body.Bytes(), &sessionResp2)
	
	coSimBody := map[string]interface{}{
		"components": []map[string]string{
			{"type": "qemu", "session_id": sessionID},
			{"type": "renode", "session_id": sessionResp2.ID},
		},
	}
	bodyCo, _ := json.Marshal(coSimBody)
	reqCo, _ := http.NewRequest("POST", "/api/v1/cosimulation", bytes.NewBuffer(bodyCo))
	reqCo.Header.Set("Content-Type", "application/json")
	wCo := httptest.NewRecorder()
	router.ServeHTTP(wCo, reqCo)
	
	if wCo.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d. Body: %s", wCo.Code, wCo.Body.String())
	}
	var coSimResp model.CoSimSession
	json.Unmarshal(wCo.Body.Bytes(), &coSimResp)
	coSimID := coSimResp.ID

	// 4.5 Start Co-Simulation
	t.Log("Testing API: Start Co-Simulation")
	reqStart, _ := http.NewRequest("POST", "/api/v1/cosimulation/"+coSimID+"/start", nil)
	wStart := httptest.NewRecorder()
	router.ServeHTTP(wStart, reqStart)
	if wStart.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d. Body: %s", wStart.Code, wStart.Body.String())
	}

	// 5. Co-Simulation Sync Time
	t.Log("Testing API: Co-Simulation Sync Time")
	syncBody := map[string]interface{}{
		"duration_ns": 1000000, // 1ms in ns
	}
	bodySync, _ := json.Marshal(syncBody)
	reqSync, _ := http.NewRequest("POST", "/api/v1/cosimulation/"+coSimID+"/sync-time", bytes.NewBuffer(bodySync))
	reqSync.Header.Set("Content-Type", "application/json")
	wSync := httptest.NewRecorder()
	router.ServeHTTP(wSync, reqSync)
	
	if wSync.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d. Body: %s", wSync.Code, wSync.Body.String())
	}
	if mockAdp.Calls["RunForTime"] != 2 {
		t.Errorf("RunForTime called %d times via API, expected 2", mockAdp.Calls["RunForTime"])
	}
}
