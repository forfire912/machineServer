package api

import (
	"net/http"
	"time"

	"github.com/forfire912/machineServer/internal/model"
	"github.com/gin-gonic/gin"
)

// CreateCoSimRequest represents a co-simulation creation request
type CreateCoSimRequest struct {
	Components []model.CoSimComponent `json:"components" binding:"required"`
}

// CreateCoSimSession godoc
// @Summary Create a new co-simulation session
// @Tags cosimulation
// @Accept json
// @Produce json
// @Param request body CreateCoSimRequest true "Co-simulation creation request"
// @Success 201 {object} model.CoSimSession
// @Router /api/v1/cosimulation [post]
func (h *Handler) CreateCoSimSession(c *gin.Context) {
	var req CreateCoSimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.service.CreateCoSimSession(c.Request.Context(), req.Components)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// StartCoSimSession godoc
// @Summary Start a co-simulation session
// @Tags cosimulation
// @Param id path string true "Session ID"
// @Success 200
// @Router /api/v1/cosimulation/{id}/start [post]
func (h *Handler) StartCoSimSession(c *gin.Context) {
	sessionID := c.Param("id")
	if err := h.service.StartCoSimSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "co-simulation started"})
}

// StopCoSimSession godoc
// @Summary Stop a co-simulation session
// @Tags cosimulation
// @Param id path string true "Session ID"
// @Success 200
// @Router /api/v1/cosimulation/{id}/stop [post]
func (h *Handler) StopCoSimSession(c *gin.Context) {
	sessionID := c.Param("id")
	if err := h.service.StopCoSimSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "co-simulation stopped"})
}

// SyncStepRequest represents a sync step request
type SyncStepRequest struct {
	Steps int `json:"steps" binding:"required"`
}

// SyncStep godoc
// @Summary Execute synchronized steps
// @Tags cosimulation
// @Param id path string true "Session ID"
// @Param request body SyncStepRequest true "Sync step request"
// @Success 200
// @Router /api/v1/cosimulation/{id}/sync-step [post]
func (h *Handler) SyncStep(c *gin.Context) {
	sessionID := c.Param("id")
	var req SyncStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SyncStep(c.Request.Context(), sessionID, req.Steps); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "steps executed"})
}

// GetCoSimSession godoc
// @Summary Get co-simulation session details
// @Tags cosimulation
// @Param id path string true "Session ID"
// @Success 200 {object} model.CoSimSession
// @Router /api/v1/cosimulation/{id} [get]
func (h *Handler) GetCoSimSession(c *gin.Context) {
	sessionID := c.Param("id")
	session, err := h.service.GetCoSimSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session)
}

// SyncTimeRequest represents a sync time request
type SyncTimeRequest struct {
	DurationNS int64 `json:"duration_ns" binding:"required"`
}

// SyncTime godoc
// @Summary Execute synchronized time slice
// @Tags cosimulation
// @Param id path string true "Session ID"
// @Param request body SyncTimeRequest true "Sync time request"
// @Success 200
// @Router /api/v1/cosimulation/{id}/sync-time [post]
func (h *Handler) SyncTime(c *gin.Context) {
	sessionID := c.Param("id")
	var req SyncTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	duration := time.Duration(req.DurationNS) * time.Nanosecond
	if err := h.service.SyncTime(c.Request.Context(), sessionID, duration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "time slice executed"})
}

// InjectEventRequest represents an event injection request
type InjectEventRequest struct {
	ComponentID string                 `json:"component_id" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Data        map[string]interface{} `json:"data" binding:"required"`
}

// InjectEvent godoc
// @Summary Inject event into co-simulation
// @Tags cosimulation
// @Param id path string true "Session ID"
// @Param request body InjectEventRequest true "Event injection request"
// @Success 200
// @Router /api/v1/cosimulation/{id}/event [post]
func (h *Handler) InjectEvent(c *gin.Context) {
	sessionID := c.Param("id")
	var req InjectEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.InjectCoSimEvent(c.Request.Context(), sessionID, req.ComponentID, req.Type, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "event injected"})
}

// ListCoSimSessions godoc
// @Summary List co-simulation sessions
// @Tags cosimulation
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cosimulation [get]
func (h *Handler) ListCoSimSessions(c *gin.Context) {
	page := 1
	pageSize := 10
	// TODO: Parse query params properly
	
	sessions, total, err := h.service.ListCoSimSessions(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data":      sessions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteCoSimSession godoc
// @Summary Delete a co-simulation session
// @Tags cosimulation
// @Param id path string true "Session ID"
// @Success 200
// @Router /api/v1/cosimulation/{id} [delete]
func (h *Handler) DeleteCoSimSession(c *gin.Context) {
	sessionID := c.Param("id")
	if err := h.service.DeleteCoSimSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "co-simulation session deleted"})
}
