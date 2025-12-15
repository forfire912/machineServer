package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/forfire912/machineServer/internal/adapter"
	"github.com/forfire912/machineServer/internal/config"
	"github.com/forfire912/machineServer/internal/model"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// EventHandler defines the interface for handling events
type EventHandler interface {
	BroadcastToSession(sessionID string, msgType string, data []byte)
}

// Service provides the core business logic
type Service struct {
	db           *gorm.DB
	config       *config.Config
	adapters     map[model.BackendType]adapter.Adapter
	sessions     map[string]*model.Session
	eventHandler EventHandler
	mu           sync.RWMutex
}

// NewService creates a new service instance
func NewService(cfg *config.Config, eventHandler EventHandler) (*Service, error) {
	// Initialize database
	db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto migrate models
	if err := db.AutoMigrate(
		&model.Session{},
		&model.Program{},
		&model.Snapshot{},
		&model.Job{},
		&model.AuditLog{},
		&model.CoSimSession{},
		&model.CoSimComponent{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize adapters
	adapters := make(map[model.BackendType]adapter.Adapter)
	if cfg.Backends.QEMU.Enabled {
		adapters[model.BackendQEMU] = adapter.NewQEMUAdapter(cfg.Backends.QEMU.Binary)
	}
	if cfg.Backends.Renode.Enabled {
		adapters[model.BackendRenode] = adapter.NewRenodeAdapter(cfg.Backends.Renode.Binary)
	}
	if cfg.Backends.OpenOCD.Enabled {
		adapters[model.BackendOpenOCD] = adapter.NewOpenOCDAdapter(cfg.Backends.OpenOCD.Binary)
	}

	// Create storage directories
	if err := os.MkdirAll(cfg.Storage.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(cfg.Storage.BasePath, "programs"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create programs directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(cfg.Storage.BasePath, "snapshots"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshots directory: %w", err)
	}
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Service{
		db:           db,
		config:       cfg,
		adapters:     adapters,
		sessions:     make(map[string]*model.Session),
		eventHandler: eventHandler,
	}, nil
}

// SetAdapter sets a backend adapter (useful for testing)
func (s *Service) SetAdapter(backend model.BackendType, adp adapter.Adapter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.adapters[backend] = adp
}

// GetCapabilities returns capabilities for all enabled backends
func (s *Service) GetCapabilities() ([]*model.Capability, error) {
	caps := make([]*model.Capability, 0, len(s.adapters))
	for _, adp := range s.adapters {
		cap, err := adp.GetCapabilities()
		if err != nil {
			return nil, err
		}
		caps = append(caps, cap)
	}
	return caps, nil
}

// CreateSession creates a new simulation session
func (s *Service) CreateSession(ctx context.Context, name string, backend model.BackendType, boardConfig *model.BoardConfig) (*model.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check session limit
	if len(s.sessions) >= s.config.Resources.MaxSessions {
		return nil, fmt.Errorf("maximum number of sessions reached")
	}

	// Get adapter
	adp, ok := s.adapters[backend]
	if !ok {
		return nil, fmt.Errorf("backend not supported: %s", backend)
	}

	// Create session
	session := &model.Session{
		ID:        uuid.New().String(),
		Name:      name,
		Backend:   backend,
		State:     model.SessionStateCreated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Marshal board config
	if boardConfig != nil {
		configJSON, err := json.Marshal(boardConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal board config: %w", err)
		}
		session.BoardConfig = string(configJSON)
	}

	// Start session with adapter
	var consoleOut io.Writer
	if s.eventHandler != nil {
		consoleOut = &consoleWriter{
			sessionID: session.ID,
			handler:   s.eventHandler,
		}
	}

	if err := adp.StartSession(ctx, session, boardConfig, consoleOut); err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	session.State = model.SessionStateRunning

	// Save to database
	if err := s.db.Create(session).Error; err != nil {
		// Try to stop the session
		adp.StopSession(ctx, session.ID)
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	s.sessions[session.ID] = session

	return session, nil
}

// consoleWriter wraps EventHandler to implement io.Writer
type consoleWriter struct {
	sessionID string
	handler   EventHandler
}

func (w *consoleWriter) Write(p []byte) (n int, err error) {
	// Make a copy of the data because p might be reused
	data := make([]byte, len(p))
	copy(data, p)
	w.handler.BroadcastToSession(w.sessionID, "console", data)
	return len(p), nil
}

// GetSession retrieves a session by ID
func (s *Service) GetSession(sessionID string) (*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if ok {
		return session, nil
	}

	// Try to load from database
	var dbSession model.Session
	if err := s.db.First(&dbSession, "id = ?", sessionID).Error; err != nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return &dbSession, nil
}

// ListSessions lists all sessions
func (s *Service) ListSessions(page, pageSize int) ([]*model.Session, int64, error) {
	var sessions []*model.Session
	var total int64

	offset := (page - 1) * pageSize

	if err := s.db.Model(&model.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// DeleteSession deletes a session
func (s *Service) DeleteSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if ok {
		// Get adapter and stop session
		if adp, exists := s.adapters[session.Backend]; exists {
			if err := adp.StopSession(ctx, sessionID); err != nil {
				return fmt.Errorf("failed to stop session: %w", err)
			}
		}
		delete(s.sessions, sessionID)
	}

	// Delete from database
	if err := s.db.Delete(&model.Session{}, "id = ?", sessionID).Error; err != nil {
		return fmt.Errorf("failed to delete session from database: %w", err)
	}

	return nil
}

// PowerOn powers on a session
func (s *Service) PowerOn(ctx context.Context, sessionID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	adp, ok := s.adapters[session.Backend]
	if !ok {
		return fmt.Errorf("adapter not found for backend: %s", session.Backend)
	}

	if err := adp.ExecuteProgram(ctx, sessionID); err != nil {
		return err
	}

	session.State = model.SessionStateRunning
	s.db.Save(session)

	return nil
}

// PowerOff powers off a session
func (s *Service) PowerOff(ctx context.Context, sessionID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	adp, ok := s.adapters[session.Backend]
	if !ok {
		return fmt.Errorf("adapter not found for backend: %s", session.Backend)
	}

	if err := adp.PauseExecution(ctx, sessionID); err != nil {
		return err
	}

	session.State = model.SessionStateStopped
	s.db.Save(session)

	return nil
}

// Reset resets a session
func (s *Service) Reset(ctx context.Context, sessionID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	adp, ok := s.adapters[session.Backend]
	if !ok {
		return fmt.Errorf("adapter not found for backend: %s", session.Backend)
	}

	return adp.ResetSession(ctx, sessionID)
}

// UploadProgram uploads a program file
func (s *Service) UploadProgram(name string, format string, reader io.Reader) (*model.Program, error) {
	programID := uuid.New().String()
	programPath := filepath.Join(s.config.Storage.BasePath, "programs", programID)

	// Create file
	file, err := os.Create(programPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create program file: %w", err)
	}
	defer file.Close()

	// Calculate hash while copying
	hash := sha256.New()
	multiWriter := io.MultiWriter(file, hash)

	size, err := io.Copy(multiWriter, reader)
	if err != nil {
		os.Remove(programPath)
		return nil, fmt.Errorf("failed to write program: %w", err)
	}

	hashStr := hex.EncodeToString(hash.Sum(nil))

	program := &model.Program{
		ID:        programID,
		Name:      name,
		Format:    format,
		Size:      size,
		Path:      programPath,
		Hash:      hashStr,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(program).Error; err != nil {
		os.Remove(programPath)
		return nil, fmt.Errorf("failed to save program: %w", err)
	}

	return program, nil
}

// LoadProgram loads a program into a session
func (s *Service) LoadProgram(ctx context.Context, sessionID, programID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	var program model.Program
	if err := s.db.First(&program, "id = ?", programID).Error; err != nil {
		return fmt.Errorf("program not found: %s", programID)
	}

	adp, ok := s.adapters[session.Backend]
	if !ok {
		return fmt.Errorf("adapter not found for backend: %s", session.Backend)
	}

	return adp.LoadProgram(ctx, sessionID, program.Path)
}

// CreateSnapshot creates a snapshot of a session
func (s *Service) CreateSnapshot(ctx context.Context, sessionID, name, description string) (*model.Snapshot, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	snapshotID := uuid.New().String()
	snapshotPath := filepath.Join(s.config.Storage.BasePath, "snapshots", snapshotID)

	adp, ok := s.adapters[session.Backend]
	if !ok {
		return nil, fmt.Errorf("adapter not found for backend: %s", session.Backend)
	}

	if err := adp.CreateSnapshot(ctx, sessionID, snapshotPath); err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	snapshot := &model.Snapshot{
		ID:          snapshotID,
		SessionID:   sessionID,
		Name:        name,
		Description: description,
		Path:        snapshotPath,
		CreatedAt:   time.Now(),
	}

	if err := s.db.Create(snapshot).Error; err != nil {
		os.Remove(snapshotPath)
		return nil, fmt.Errorf("failed to save snapshot: %w", err)
	}

	return snapshot, nil
}

// RestoreSnapshot restores a session from a snapshot
func (s *Service) RestoreSnapshot(ctx context.Context, sessionID, snapshotID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	var snapshot model.Snapshot
	if err := s.db.First(&snapshot, "id = ?", snapshotID).Error; err != nil {
		return fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	adp, ok := s.adapters[session.Backend]
	if !ok {
		return fmt.Errorf("adapter not found for backend: %s", session.Backend)
	}

	return adp.RestoreSnapshot(ctx, sessionID, snapshot.Path)
}

// LogAudit logs an audit entry
func (s *Service) LogAudit(userID, action, resource, details, ip string) error {
	log := &model.AuditLog{
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Details:   details,
		IP:        ip,
		Timestamp: time.Now(),
	}

	return s.db.Create(log).Error
}

// StartCoverage starts coverage collection
func (s *Service) StartCoverage(ctx context.Context, sessionID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	adp, ok := s.adapters[session.Backend]
	if !ok {
		return fmt.Errorf("adapter not found for backend: %s", session.Backend)
	}

	// Define output path
	// Ensure coverage directory exists
	coverageDir := filepath.Join(s.config.Storage.BasePath, "coverage")
	if err := os.MkdirAll(coverageDir, 0755); err != nil {
		return fmt.Errorf("failed to create coverage directory: %w", err)
	}

	outputPath := filepath.Join(coverageDir, fmt.Sprintf("%s.trace", sessionID))

	return adp.StartCoverage(ctx, sessionID, outputPath)
}

// StopCoverage stops coverage collection
func (s *Service) StopCoverage(ctx context.Context, sessionID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	adp, ok := s.adapters[session.Backend]
	if !ok {
		return fmt.Errorf("adapter not found for backend: %s", session.Backend)
	}

	return adp.StopCoverage(ctx, sessionID)
}
