package api

import (
	"github.com/forfire912/machineServer/internal/config"
	"github.com/forfire912/machineServer/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the HTTP router with all routes
func SetupRouter(cfg *config.Config, svc *service.Service) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	router := gin.Default()

	// Apply global middleware
	router.Use(CORSMiddleware())
	router.Use(AuditMiddleware(svc))
	router.Use(PrometheusMiddleware())

	// Create stream hub
	streamHub := NewStreamHub()
	go streamHub.Run()

	// Create handler
	handler := NewHandler(svc, streamHub)

	// Health check (no auth required)
	router.GET("/health", handler.HealthCheck)

	// Metrics endpoint (no auth required)
	if cfg.Monitoring.Enabled {
		router.GET("/metrics", PrometheusHandler())
	}

	// API v1 routes (with auth)
	v1 := router.Group("/api/v1")
	v1.Use(AuthMiddleware(cfg))
	{
		// Capabilities
		v1.GET("/capabilities", handler.GetCapabilities)

		// Sessions
		sessions := v1.Group("/sessions")
		{
			sessions.POST("", handler.CreateSession)
			sessions.GET("", handler.ListSessions)
			sessions.GET("/:id", handler.GetSession)
			sessions.DELETE("/:id", handler.DeleteSession)

			// Session control
			sessions.POST("/:id/poweron", handler.PowerOn)
			sessions.POST("/:id/poweroff", handler.PowerOff)
			sessions.POST("/:id/reset", handler.Reset)

			// Program management
			sessions.POST("/:id/program", handler.LoadProgram)

			// Snapshots
			sessions.POST("/:id/snapshots", handler.CreateSnapshot)
			sessions.POST("/:id/restore", handler.RestoreSnapshot)

			// WebSocket streams
			sessions.GET("/:id/stream/console", handler.StreamConsole)
		}

		// Programs
		programs := v1.Group("/programs")
		{
			programs.POST("", handler.UploadProgram)
		}
	}

	return router
}
