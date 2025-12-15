package api

import (
	"net/http"
	"strconv"

	"github.com/forfire912/machineServer/internal/model"
	"github.com/forfire912/machineServer/internal/service"
	"github.com/gin-gonic/gin"
)

// Handler provides HTTP API handlers
type Handler struct {
	service   *service.Service
	streamHub *StreamHub
}

// NewHandler creates a new API handler
func NewHandler(svc *service.Service, hub *StreamHub) *Handler {
	return &Handler{
		service:   svc,
		streamHub: hub,
	}
}

// GetCapabilities godoc
// @Summary Get backend capabilities
// @Description Get capabilities of all enabled simulation backends
// @Tags capabilities
// @Produce json
// @Success 200 {array} model.Capability
// @Router /api/v1/capabilities [get]
func (h *Handler) GetCapabilities(c *gin.Context) {
	caps, err := h.service.GetCapabilities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, caps)
}

// CreateSessionRequest represents a session creation request
type CreateSessionRequest struct {
	Name        string               `json:"name" binding:"required"`
	Backend     model.BackendType    `json:"backend" binding:"required"`
	BoardConfig *model.BoardConfig   `json:"board_config,omitempty"`
}

// CreateSession godoc
// @Summary Create a new simulation session
// @Description Create a new simulation session with specified backend and configuration
// @Tags sessions
// @Accept json
// @Produce json
// @Param request body CreateSessionRequest true "Session creation request"
// @Success 201 {object} model.Session
// @Router /api/v1/sessions [post]
func (h *Handler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.service.CreateSession(c.Request.Context(), req.Name, req.Backend, req.BoardConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GetSession godoc
// @Summary Get session details
// @Description Get detailed information about a specific session
// @Tags sessions
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} model.Session
// @Router /api/v1/sessions/{id} [get]
func (h *Handler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")

	session, err := h.service.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// ListSessions godoc
// @Summary List all sessions
// @Description List all simulation sessions with pagination
// @Tags sessions
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	sessions, total, err := h.service.ListSessions(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions":  sessions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteSession godoc
// @Summary Delete a session
// @Description Delete a simulation session and clean up resources
// @Tags sessions
// @Param id path string true "Session ID"
// @Success 204
// @Router /api/v1/sessions/{id} [delete]
func (h *Handler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")

	if err := h.service.DeleteSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// PowerOn godoc
// @Summary Power on a session
// @Description Start execution in a simulation session
// @Tags control
// @Param id path string true "Session ID"
// @Success 200
// @Router /api/v1/sessions/{id}/poweron [post]
func (h *Handler) PowerOn(c *gin.Context) {
	sessionID := c.Param("id")

	if err := h.service.PowerOn(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session powered on"})
}

// PowerOff godoc
// @Summary Power off a session
// @Description Stop execution in a simulation session
// @Tags control
// @Param id path string true "Session ID"
// @Success 200
// @Router /api/v1/sessions/{id}/poweroff [post]
func (h *Handler) PowerOff(c *gin.Context) {
	sessionID := c.Param("id")

	if err := h.service.PowerOff(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session powered off"})
}

// Reset godoc
// @Summary Reset a session
// @Description Reset a simulation session to initial state
// @Tags control
// @Param id path string true "Session ID"
// @Success 200
// @Router /api/v1/sessions/{id}/reset [post]
func (h *Handler) Reset(c *gin.Context) {
	sessionID := c.Param("id")

	if err := h.service.Reset(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session reset"})
}

// UploadProgram godoc
// @Summary Upload a program
// @Description Upload a program file (ELF, Binary, or HEX format)
// @Tags programs
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Program file"
// @Param name formData string true "Program name"
// @Param format formData string true "Program format (elf, binary, hex)"
// @Success 201 {object} model.Program
// @Router /api/v1/programs [post]
func (h *Handler) UploadProgram(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	name := c.PostForm("name")
	format := c.PostForm("format")

	if name == "" {
		name = file.Filename
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	program, err := h.service.UploadProgram(name, format, src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, program)
}

// LoadProgramRequest represents a program load request
type LoadProgramRequest struct {
	ProgramID string `json:"program_id" binding:"required"`
}

// LoadProgram godoc
// @Summary Load a program into a session
// @Description Load a previously uploaded program into a simulation session
// @Tags programs
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param request body LoadProgramRequest true "Program load request"
// @Success 200
// @Router /api/v1/sessions/{id}/program [post]
func (h *Handler) LoadProgram(c *gin.Context) {
	sessionID := c.Param("id")

	var req LoadProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.LoadProgram(c.Request.Context(), sessionID, req.ProgramID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "program loaded"})
}

// CreateSnapshotRequest represents a snapshot creation request
type CreateSnapshotRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
}

// CreateSnapshot godoc
// @Summary Create a snapshot
// @Description Create a snapshot of the current simulation state
// @Tags snapshots
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param request body CreateSnapshotRequest true "Snapshot creation request"
// @Success 201 {object} model.Snapshot
// @Router /api/v1/sessions/{id}/snapshots [post]
func (h *Handler) CreateSnapshot(c *gin.Context) {
	sessionID := c.Param("id")

	var req CreateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	snapshot, err := h.service.CreateSnapshot(c.Request.Context(), sessionID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, snapshot)
}

// RestoreSnapshotRequest represents a snapshot restore request
type RestoreSnapshotRequest struct {
	SnapshotID string `json:"snapshot_id" binding:"required"`
}

// RestoreSnapshot godoc
// @Summary Restore from a snapshot
// @Description Restore a simulation session to a previous snapshot
// @Tags snapshots
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param request body RestoreSnapshotRequest true "Snapshot restore request"
// @Success 200
// @Router /api/v1/sessions/{id}/restore [post]
func (h *Handler) RestoreSnapshot(c *gin.Context) {
	sessionID := c.Param("id")

	var req RestoreSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.RestoreSnapshot(c.Request.Context(), sessionID, req.SnapshotID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "snapshot restored"})
}

// HealthCheck godoc
// @Summary Health check
// @Description Check if the service is running
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "1.0.0",
	})
}
