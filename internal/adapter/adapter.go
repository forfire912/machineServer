package adapter

import (
	"context"
	"io"
	"time"

	"github.com/forfire912/machineServer/internal/model"
)

// Adapter defines the interface for simulation backend adapters
type Adapter interface {
	// GetCapabilities returns the capabilities of the backend
	GetCapabilities() (*model.Capability, error)

	// StartSession starts a new simulation session
	StartSession(ctx context.Context, session *model.Session, config *model.BoardConfig, consoleOut io.Writer) error

	// StopSession stops a running session
	StopSession(ctx context.Context, sessionID string) error

	// ResetSession resets a session
	ResetSession(ctx context.Context, sessionID string) error

	// LoadProgram loads a program into the simulation
	LoadProgram(ctx context.Context, sessionID string, programPath string) error

	// ExecuteProgram starts program execution
	ExecuteProgram(ctx context.Context, sessionID string) error

	// PauseExecution pauses program execution
	PauseExecution(ctx context.Context, sessionID string) error

	// ResumeExecution resumes program execution
	ResumeExecution(ctx context.Context, sessionID string) error

	// GetGDBPort returns the GDB server port
	GetGDBPort(sessionID string) (int, error)

	// CreateSnapshot creates a snapshot of the current state
	CreateSnapshot(ctx context.Context, sessionID string, snapshotPath string) error

	// RestoreSnapshot restores from a snapshot
	RestoreSnapshot(ctx context.Context, sessionID string, snapshotPath string) error

	// GetConsoleOutput gets console output
	GetConsoleOutput(ctx context.Context, sessionID string) (string, error)

	// Step executes a number of instructions
	Step(ctx context.Context, sessionID string, steps int) error

	// StartCoverage starts coverage collection
	StartCoverage(ctx context.Context, sessionID string, outputPath string) error

	// StopCoverage stops coverage collection
	StopCoverage(ctx context.Context, sessionID string) error

	// RunForTime runs the simulation for a specific duration (Scheme 3)
	RunForTime(ctx context.Context, sessionID string, duration time.Duration) error

	// InjectEvent injects an event into the simulation (Scheme 4)
	InjectEvent(ctx context.Context, sessionID string, eventType string, data map[string]interface{}) error
}
