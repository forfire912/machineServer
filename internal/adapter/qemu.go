package adapter

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

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
			// ARM Cortex-M
			"cortex-m0", "cortex-m3", "cortex-m4", "cortex-m7", "cortex-m33",
			// ARM Cortex-A
			"cortex-a7", "cortex-a8", "cortex-a9", "cortex-a15", "cortex-a53", "cortex-a57", "cortex-a72",
			// ARM Classic
			"arm926", "arm1136",
			// RISC-V
			"riscv32", "riscv64", "sifive-e31", "sifive-u54",
			// x86
			"i386", "x86_64",
		},
		Peripherals: []string{
			"uart", "pl011", "16550a",
			"gpio", "pl061",
			"spi", "ssi",
			"i2c",
			"timer", "sp804", "arm_timer",
			"adc",
			"ethernet", "smc91c111", "lan9118", "e1000", "virtio-net",
			"display", "pl110",
			"sd", "pl181", "sdhci",
			"usb", "usb-ehci", "usb-ohci",
			"virtio-blk", "virtio-rng",
		},
		BusTypes: []string{
			"ahb", "apb", "axi",
			"pci", "pcie",
			"usb",
			"i2c", "spi",
		},
		Features: []string{
			"gdb-server",
			"snapshot",
			"monitor",
			"qmp",
		},
		Boards: []string{
			// ARM
			"versatilepb", "vexpress-a9", "realview-eb", "integratorcp",
			"mps2-an385", "mps2-an500", "mps2-an511",
			"stm32vldiscovery", "stm32f405soc", "netduino2", "netduinoplus2",
			"microbit", "nrf51dk",
			"raspi2", "raspi3",
			// RISC-V
			"virt", "sifive_e", "sifive_u", "spike",
			// x86
			"pc", "q35", "isapc",
		},
	}, nil
}

func (q *QEMUAdapter) StartSession(ctx context.Context, session *model.Session, config *model.BoardConfig, consoleOut io.Writer) error {
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
		"-nographic",
		"-s", // GDB server
		"-gdb", fmt.Sprintf("tcp::%d", gdbPort),
		"-monitor", fmt.Sprintf("tcp:127.0.0.1:%d,server,nowait", monitorPort),
		"-S", // Start paused
		"-semihosting-config", "enable=on,target=native", // Enable semihosting for coverage data
	}

	// Configure Machine/Board
	machine := "netduino2" // Default
	if config != nil {
		if config.Board != "" {
			machine = config.Board
		} else if config.Processor.Model != "" {
			// Try to map processor to a generic machine if possible
			switch config.Processor.Model {
			case "cortex-m3":
				machine = "lm3s6965evb"
			case "cortex-m4":
				machine = "netduino2"
			case "riscv32", "riscv64":
				machine = "virt"
			case "aarch64", "cortex-a53", "cortex-a57":
				machine = "virt"
			}
		}
	}
	args = append(args, "-M", machine)

	// Configure CPU if using generic machine
	if config != nil && config.Processor.Model != "" && (machine == "virt" || machine == "versatilepb") {
		args = append(args, "-cpu", config.Processor.Model)
	}

	// Configure Memory
	if config != nil && config.Memory.RAM.Size > 0 {
		args = append(args, "-m", fmt.Sprintf("%dM", config.Memory.RAM.Size/(1024*1024)))
	}

	// Configure Peripherals (Network, etc.)
	if config != nil {
		for i, p := range config.Peripherals {
			if p.Type == "ethernet" || p.Type == "virtio-net" {
				// Add user networking
				netID := fmt.Sprintf("net%d", i)
				args = append(args, "-netdev", fmt.Sprintf("user,id=%s", netID))
				args = append(args, "-device", fmt.Sprintf("virtio-net-device,netdev=%s", netID))
			}
		}
	}

	cmd := exec.CommandContext(ctx, q.binaryPath, args...)
	
	// Set working directory to a temp dir to capture semihosting output (e.g. gcda files)
	workDir := filepath.Join(os.TempDir(), "qemu_session_"+session.ID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create session work dir: %w", err)
	}
	cmd.Dir = workDir

	if consoleOut != nil {
		cmd.Stdout = consoleOut
		cmd.Stderr = consoleOut
	}

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
	q.mu.RLock()
	sess, ok := q.sessions[sessionID]
	q.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Create GDB script
	scriptContent := fmt.Sprintf("target remote localhost:%d\nload %s\nquit\n", sess.gdbPort, programPath)
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("load_%s.gdb", sessionID))
	if err := os.WriteFile(tmpFile, []byte(scriptContent), 0644); err != nil {
		return fmt.Errorf("failed to create GDB script: %w", err)
	}
	defer os.Remove(tmpFile)

	// Run GDB
	cmd := exec.CommandContext(ctx, "gdb", "-batch", "-x", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to load program via GDB: %w, output: %s", err, string(output))
	}

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

func (q *QEMUAdapter) Step(ctx context.Context, sessionID string, steps int) error {
	// QEMU monitor doesn't support simple 'step' command easily without GDB.
	// For now, we'll return not supported or implement via GDB later.
	return fmt.Errorf("step not supported for QEMU backend without GDB connection")
}

func (q *QEMUAdapter) StartCoverage(ctx context.Context, sessionID string, outputPath string) error {
	// For QEMU, coverage is collected by the firmware via semihosting.
	// We just need to ensure the environment is ready (which is done in StartSession via -semihosting-config).
	// We can log that we are expecting coverage data.
	return nil
}

func (q *QEMUAdapter) StopCoverage(ctx context.Context, sessionID string) error {
	// In a real implementation, we might want to scan the workDir for .gcda files 
	// and move them to the expected storage location.
	// For now, we assume the user/tools will retrieve them from the session artifact directory.
	return nil
}

func (q *QEMUAdapter) RunForTime(ctx context.Context, sessionID string, duration time.Duration) error {
	// QEMU doesn't support "run for time" natively in monitor.
	// Scheme 3 (Soft Real-time) approximation:
	// Resume -> Sleep(duration) -> Stop
	
	if err := q.ResumeExecution(ctx, sessionID); err != nil {
		return err
	}
	
	select {
	case <-time.After(duration):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	return q.PauseExecution(ctx, sessionID)
}

func (q *QEMUAdapter) InjectEvent(ctx context.Context, sessionID string, eventType string, data map[string]interface{}) error {
	// Scheme 4: Event Injection via Monitor
	// Example: send_key, mouse_move, etc.
	
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

	var cmd string
	switch eventType {
	case "key":
		// data: {"key": "ctrl-c"}
		key, _ := data["key"].(string)
		cmd = fmt.Sprintf("sendkey %s\n", key)
	case "mouse_move":
		// data: {"dx": 10, "dy": 20}
		dx, _ := data["dx"].(int)
		dy, _ := data["dy"].(int)
		cmd = fmt.Sprintf("mouse_move %d %d\n", dx, dy)
	default:
		return fmt.Errorf("unsupported event type: %s", eventType)
	}

	_, err = conn.Write([]byte(cmd))
	return err
}
