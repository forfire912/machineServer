package api

import (
	"github.com/forfire912/machineServer/internal/config"
	"github.com/forfire912/machineServer/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the HTTP router with all routes
func SetupRouter(cfg *config.Config, svc *service.Service, streamHub *StreamHub) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	router := gin.Default()

	// Apply global middleware
	router.Use(CORSMiddleware())
	router.Use(AuditMiddleware(svc))
	router.Use(PrometheusMiddleware())

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

			// Coverage
			sessions.POST("/:id/coverage/start", handler.StartCoverage)
			sessions.POST("/:id/coverage/stop", handler.StopCoverage)

			// WebSocket streams
			sessions.GET("/:id/stream/console", handler.StreamConsole)
		}

		// Programs
		programs := v1.Group("/programs")
		{
			programs.POST("", handler.UploadProgram)
		}

		// Co-Simulation
		cosim := v1.Group("/cosimulation")
		{
			cosim.POST("", handler.CreateCoSimSession)
			cosim.GET("", handler.ListCoSimSessions)
			cosim.GET("/:id", handler.GetCoSimSession)
			cosim.DELETE("/:id", handler.DeleteCoSimSession)
			cosim.POST("/:id/start", handler.StartCoSimSession)
			cosim.POST("/:id/stop", handler.StopCoSimSession)
			cosim.POST("/:id/sync-step", handler.SyncStep)
			cosim.POST("/:id/sync-time", handler.SyncTime)
			cosim.POST("/:id/event", handler.InjectEvent)
		}
	}

	return router
}
