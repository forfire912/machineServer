package adapter

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/forfire912/machineServer/internal/model"
)

// RenodeAdapter implements the Adapter interface for Renode
type RenodeAdapter struct {
	binaryPath string
	sessions   map[string]*renodeSession
	mu         sync.RWMutex
}

type renodeSession struct {
	cmd     *exec.Cmd
	gdbPort int
	telnetPort int
}

// NewRenodeAdapter creates a new Renode adapter
func NewRenodeAdapter(binaryPath string) *RenodeAdapter {
	return &RenodeAdapter{
		binaryPath: binaryPath,
		sessions:   make(map[string]*renodeSession),
	}
}

func (r *RenodeAdapter) GetCapabilities() (*model.Capability, error) {
	return &model.Capability{
		Backend: model.BackendRenode,
		Processors: []string{
			"cortex-m3",
			"cortex-m4",
			"cortex-m7",
			"cortex-m33",
			"cortex-a9",
			"riscv32",
			"riscv64",
		},
		Peripherals: []string{
			"uart",
			"gpio",
			"spi",
			"i2c",
			"timer",
			"adc",
			"can",
			"ethernet",
		},
		BusTypes: []string{
			"ahb",
			"apb",
			"axi",
		},
		Features: []string{
			"gdb-server",
			"snapshot",
			"monitor",
			"multi-node",
			"time-sync",
		},
	}, nil
}

func (r *RenodeAdapter) StartSession(ctx context.Context, session *model.Session, config *model.BoardConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gdbPort, err := getFreePort()
	if err != nil {
		return fmt.Errorf("failed to get GDB port: %w", err)
	}

	telnetPort, err := getFreePort()
	if err != nil {
		return fmt.Errorf("failed to get telnet port: %w", err)
	}

	// Build Renode command
	args := []string{
		"--disable-gui",
		"--port", fmt.Sprintf("%d", telnetPort),
	}

	cmd := exec.CommandContext(ctx, r.binaryPath, args...)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Renode: %w", err)
	}

	r.sessions[session.ID] = &renodeSession{
		cmd:        cmd,
		gdbPort:    gdbPort,
		telnetPort: telnetPort,
	}

	session.GDBPort = gdbPort
	session.MonitorPort = telnetPort
	session.PID = cmd.Process.Pid

	return nil
}

func (r *RenodeAdapter) StopSession(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sess, ok := r.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if sess.cmd != nil && sess.cmd.Process != nil {
		if err := sess.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill Renode process: %w", err)
		}
	}

	delete(r.sessions, sessionID)
	return nil
}

func (r *RenodeAdapter) ResetSession(ctx context.Context, sessionID string) error {
	// Send reset command via telnet connection
	return nil
}

func (r *RenodeAdapter) LoadProgram(ctx context.Context, sessionID string, programPath string) error {
	// Load program via Renode monitor commands
	return nil
}

func (r *RenodeAdapter) ExecuteProgram(ctx context.Context, sessionID string) error {
	// Send start command
	return nil
}

func (r *RenodeAdapter) PauseExecution(ctx context.Context, sessionID string) error {
	// Send pause command
	return nil
}

func (r *RenodeAdapter) ResumeExecution(ctx context.Context, sessionID string) error {
	// Send resume command
	return nil
}

func (r *RenodeAdapter) GetGDBPort(sessionID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sess, ok := r.sessions[sessionID]
	if !ok {
		return 0, fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.gdbPort, nil
}

func (r *RenodeAdapter) CreateSnapshot(ctx context.Context, sessionID string, snapshotPath string) error {
	// Create snapshot via Renode commands
	return nil
}

func (r *RenodeAdapter) RestoreSnapshot(ctx context.Context, sessionID string, snapshotPath string) error {
	// Restore snapshot via Renode commands
	return nil
}

func (r *RenodeAdapter) GetConsoleOutput(ctx context.Context, sessionID string) (string, error) {
	return "", nil
}
