package adapter

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"sync"

	"github.com/forfire912/machineServer/internal/model"
)

// QEMUAdapter implements the Adapter interface for QEMU
type QEMUAdapter struct {
	binaryPath string
	sessions   map[string]*qemuSession
	mu         sync.RWMutex
}

type qemuSession struct {
	cmd         *exec.Cmd
	gdbPort     int
	monitorPort int
	qmpConn     net.Conn
}

// NewQEMUAdapter creates a new QEMU adapter
func NewQEMUAdapter(binaryPath string) *QEMUAdapter {
	return &QEMUAdapter{
		binaryPath: binaryPath,
		sessions:   make(map[string]*qemuSession),
	}
}

func (q *QEMUAdapter) GetCapabilities() (*model.Capability, error) {
	return &model.Capability{
		Backend: model.BackendQEMU,
		Processors: []string{
			"cortex-m3",
			"cortex-m4",
			"cortex-m7",
			"cortex-m33",
			"arm926",
		},
		Peripherals: []string{
			"uart",
			"gpio",
			"spi",
			"i2c",
			"timer",
			"adc",
		},
		BusTypes: []string{
			"ahb",
			"apb",
		},
		Features: []string{
			"gdb-server",
			"snapshot",
			"monitor",
			"qmp",
		},
	}, nil
}

func (q *QEMUAdapter) StartSession(ctx context.Context, session *model.Session, config *model.BoardConfig) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Find free ports
	gdbPort, err := getFreePort()
	if err != nil {
		return fmt.Errorf("failed to get GDB port: %w", err)
	}

	monitorPort, err := getFreePort()
	if err != nil {
		return fmt.Errorf("failed to get monitor port: %w", err)
	}

	// Build QEMU command
	args := []string{
		"-M", "netduino2", // Default machine
		"-nographic",
		"-s", // GDB server on port 1234
		"-gdb", fmt.Sprintf("tcp::%d", gdbPort),
		"-monitor", fmt.Sprintf("tcp:127.0.0.1:%d,server,nowait", monitorPort),
		"-S", // Start paused
	}

	// Add memory configuration
	if config != nil {
		args = append(args, "-m", fmt.Sprintf("%dM", config.Memory.RAM.Size/(1024*1024)))
	}

	cmd := exec.CommandContext(ctx, q.binaryPath, args...)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start QEMU: %w", err)
	}

	q.sessions[session.ID] = &qemuSession{
		cmd:         cmd,
		gdbPort:     gdbPort,
		monitorPort: monitorPort,
	}

	session.GDBPort = gdbPort
	session.MonitorPort = monitorPort
	session.PID = cmd.Process.Pid

	return nil
}

func (q *QEMUAdapter) StopSession(ctx context.Context, sessionID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	sess, ok := q.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if sess.cmd != nil && sess.cmd.Process != nil {
		if err := sess.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill QEMU process: %w", err)
		}
	}

	delete(q.sessions, sessionID)
	return nil
}

func (q *QEMUAdapter) ResetSession(ctx context.Context, sessionID string) error {
	q.mu.RLock()
	sess, ok := q.sessions[sessionID]
	q.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Send reset command via monitor
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", sess.monitorPort))
	if err != nil {
		return fmt.Errorf("failed to connect to monitor: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("system_reset\n"))
	return err
}

func (q *QEMUAdapter) LoadProgram(ctx context.Context, sessionID string, programPath string) error {
	// Program loading is typically done via GDB or command line
	// This is a placeholder implementation
	return nil
}

func (q *QEMUAdapter) ExecuteProgram(ctx context.Context, sessionID string) error {
	q.mu.RLock()
	sess, ok := q.sessions[sessionID]
	q.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Send continue command via monitor
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", sess.monitorPort))
	if err != nil {
		return fmt.Errorf("failed to connect to monitor: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("cont\n"))
	return err
}

func (q *QEMUAdapter) PauseExecution(ctx context.Context, sessionID string) error {
	q.mu.RLock()
	sess, ok := q.sessions[sessionID]
	q.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", sess.monitorPort))
	if err != nil {
		return fmt.Errorf("failed to connect to monitor: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("stop\n"))
	return err
}

func (q *QEMUAdapter) ResumeExecution(ctx context.Context, sessionID string) error {
	return q.ExecuteProgram(ctx, sessionID)
}

func (q *QEMUAdapter) GetGDBPort(sessionID string) (int, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	sess, ok := q.sessions[sessionID]
	if !ok {
		return 0, fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.gdbPort, nil
}

func (q *QEMUAdapter) CreateSnapshot(ctx context.Context, sessionID string, snapshotPath string) error {
	q.mu.RLock()
	sess, ok := q.sessions[sessionID]
	q.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", sess.monitorPort))
	if err != nil {
		return fmt.Errorf("failed to connect to monitor: %w", err)
	}
	defer conn.Close()

	cmd := fmt.Sprintf("savevm %s\n", snapshotPath)
	_, err = conn.Write([]byte(cmd))
	return err
}

func (q *QEMUAdapter) RestoreSnapshot(ctx context.Context, sessionID string, snapshotPath string) error {
	q.mu.RLock()
	sess, ok := q.sessions[sessionID]
	q.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", sess.monitorPort))
	if err != nil {
		return fmt.Errorf("failed to connect to monitor: %w", err)
	}
	defer conn.Close()

	cmd := fmt.Sprintf("loadvm %s\n", snapshotPath)
	_, err = conn.Write([]byte(cmd))
	return err
}

func (q *QEMUAdapter) GetConsoleOutput(ctx context.Context, sessionID string) (string, error) {
	// This would require capturing stdout/stderr from the QEMU process
	return "", nil
}

// Helper function to get a free port
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
