package service

import (
	"context"
	"fmt"
	"time"

	"github.com/forfire912/machineServer/internal/model"
	"github.com/google/uuid"
)

// CreateCoSimSession creates a new co-simulation session
func (s *Service) CreateCoSimSession(ctx context.Context, components []model.CoSimComponent) (*model.CoSimSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := "cosim_" + uuid.New().String()[:8]
	session := &model.CoSimSession{
		ID:        sessionID,
		Status:    "created",
		CreatedAt: time.Now(),
		Components: components,
	}

	// Initialize components
	for i := range session.Components {
		session.Components[i].ID = "comp_" + uuid.New().String()[:8]
		session.Components[i].CoSimID = sessionID
		session.Components[i].Status = "initialized"
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, err
	}

	return session, nil
}

// StartCoSimSession starts the co-simulation
func (s *Service) StartCoSimSession(ctx context.Context, sessionID string) error {
	now := time.Now()
	return s.db.Model(&model.CoSimSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"status":     "running",
		"started_at": &now,
	}).Error
}

// StopCoSimSession stops the co-simulation
func (s *Service) StopCoSimSession(ctx context.Context, sessionID string) error {
	return s.db.Model(&model.CoSimSession{}).Where("id = ?", sessionID).Update("status", "stopped").Error
}

// SyncStep executes a synchronized step for all components
func (s *Service) SyncStep(ctx context.Context, sessionID string, steps int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get session
	var session model.CoSimSession
	if err := s.db.Preload("Components").First(&session, "id = ?", sessionID).Error; err != nil {
		return err
	}

	if session.Status != "running" {
		return fmt.Errorf("session is not running")
	}

	// Step all components
	for _, comp := range session.Components {
		if comp.SessionID != "" {
			// Find the adapter for this component's session
			sess, ok := s.sessions[comp.SessionID]
			if ok {
				if adp, ok := s.adapters[sess.Backend]; ok {
					// Ignore errors for now as some backends might not support stepping
					_ = adp.Step(ctx, sess.ID, steps)
				}
			}
		}
	}

	session.SyncCount += int64(steps)
	session.TimeNS += int64(steps * 1000) // Assuming 1us per step for now

	return s.db.Save(&session).Error
}

// GetCoSimSession retrieves a co-simulation session
func (s *Service) GetCoSimSession(sessionID string) (*model.CoSimSession, error) {
	var session model.CoSimSession
	if err := s.db.Preload("Components").First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// ListCoSimSessions lists all co-simulation sessions
func (s *Service) ListCoSimSessions(page, pageSize int) ([]*model.CoSimSession, int64, error) {
	var sessions []*model.CoSimSession
	var total int64

	offset := (page - 1) * pageSize

	if err := s.db.Model(&model.CoSimSession{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Preload("Components").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// DeleteCoSimSession deletes a co-simulation session
func (s *Service) DeleteCoSimSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session exists
	var session model.CoSimSession
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return err
	}

	// Stop if running (best effort)
	if session.Status == "running" {
		_ = s.db.Model(&model.CoSimSession{}).Where("id = ?", sessionID).Update("status", "stopped").Error
	}

	// Delete components
	if err := s.db.Delete(&model.CoSimComponent{}, "co_sim_id = ?", sessionID).Error; err != nil {
		return fmt.Errorf("failed to delete session components: %w", err)
	}

	// Delete session
	if err := s.db.Delete(&model.CoSimSession{}, "id = ?", sessionID).Error; err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// SyncTime executes a synchronized time slice for all components (Scheme 3)
func (s *Service) SyncTime(ctx context.Context, sessionID string, duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get session
	var session model.CoSimSession
	if err := s.db.Preload("Components").First(&session, "id = ?", sessionID).Error; err != nil {
		return err
	}

	if session.Status != "running" {
		return fmt.Errorf("session is not running")
	}

	// Run all components for duration
	// Note: Ideally this should be parallel, but for simplicity we do serial start -> wait -> stop logic inside adapter
	// or we launch goroutines.
	// Given the Adapter.RunForTime implementation (blocking), serial execution means total time = N * duration, which is wrong.
	// We need to launch them in parallel.
	
	errChan := make(chan error, len(session.Components))
	for _, comp := range session.Components {
		if comp.SessionID != "" {
			go func(compID, sessID string) {
				sess, ok := s.sessions[sessID]
				if ok {
					if adp, ok := s.adapters[sess.Backend]; ok {
						errChan <- adp.RunForTime(ctx, sess.ID, duration)
						return
					}
				}
				errChan <- nil
			}(comp.ID, comp.SessionID)
		} else {
			errChan <- nil
		}
	}

	// Wait for all
	for i := 0; i < len(session.Components); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	session.TimeNS += int64(duration)
	return s.db.Save(&session).Error
}

// InjectCoSimEvent injects an event into a specific component (Scheme 4)
func (s *Service) InjectCoSimEvent(ctx context.Context, sessionID string, componentID string, eventType string, data map[string]interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get session to verify component belongs to it
	var session model.CoSimSession
	if err := s.db.Preload("Components").First(&session, "id = ?", sessionID).Error; err != nil {
		return err
	}

	// Find component
	var targetComp *model.CoSimComponent
	for _, comp := range session.Components {
		if comp.ID == componentID {
			targetComp = &comp
			break
		}
	}
	if targetComp == nil {
		return fmt.Errorf("component not found: %s", componentID)
	}

	if targetComp.SessionID == "" {
		return fmt.Errorf("component not initialized")
	}

	sess, ok := s.sessions[targetComp.SessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}

	adp, ok := s.adapters[sess.Backend]
	if !ok {
		return fmt.Errorf("adapter not found")
	}

	return adp.InjectEvent(ctx, sess.ID, eventType, data)
}
