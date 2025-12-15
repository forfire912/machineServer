package adapter

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"sync"
	"time"

	"github.com/forfire912/machineServer/internal/model"
)

// OpenOCDAdapter implements the Adapter interface for OpenOCD
type OpenOCDAdapter struct {
	binaryPath string
	sessions   map[string]*openocdSession
	mu         sync.RWMutex
}

type openocdSession struct {
	cmd        *exec.Cmd
	gdbPort    int
	telnetPort int
}

// NewOpenOCDAdapter creates a new OpenOCD adapter
func NewOpenOCDAdapter(binaryPath string) *OpenOCDAdapter {
	return &OpenOCDAdapter{
		binaryPath: binaryPath,
		sessions:   make(map[string]*openocdSession),
	}
}

func (o *OpenOCDAdapter) GetCapabilities() (*model.Capability, error) {
	return &model.Capability{
		Backend: model.BackendOpenOCD,
		Processors: []string{
			"cortex-m3",
			"cortex-m4",
			"cortex-m7",
			"cortex-m33",
			"stm32f1x",
			"stm32f4x",
		},
		Peripherals: []string{
			"hardware-dependent",
		},
		BusTypes: []string{
			"jtag",
			"swd",
		},
		Features: []string{
			"gdb-server",
			"flash-programming",
			"reset-control",
		},
		Boards: []string{
			// ST
			"st_nucleo_f103rb", "st_nucleo_f4", "stm32f3discovery", "stm32f4discovery", "stm32f7discovery",
			// NXP
			"frdm-k64f", "imxrt1050-evk",
			// Nordic
			"nrf51dk", "nrf52dk",
			// TI
			"ek-tm4c123gxl", "ek-tm4c1294xl",
			// Raspberry Pi
			"rpi_pico",
			// Generic
			"generic_board",
		},
	}, nil
}

func (o *OpenOCDAdapter) StartSession(ctx context.Context, session *model.Session, config *model.BoardConfig, consoleOut io.Writer) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	gdbPort, err := getFreePort()
	if err != nil {
		return fmt.Errorf("failed to get GDB port: %w", err)
	}

	telnetPort, err := getFreePort()
	if err != nil {
		return fmt.Errorf("failed to get telnet port: %w", err)
	}

	// Build OpenOCD command
	// Note: In a real implementation, we would map config.Processor.Model to specific config files
	// For now, we use some defaults or placeholders based on the model
	
	interfaceCfg := "interface/stlink.cfg" // Default to ST-Link
	targetCfg := "target/stm32f4x.cfg"     // Default to STM32F4
	
	if config != nil && config.Processor.Model != "" {
		// Simple mapping logic
		switch config.Processor.Model {
		case "stm32f1x":
			targetCfg = "target/stm32f1x.cfg"
		case "stm32f4x":
			targetCfg = "target/stm32f4x.cfg"
		case "stm32h7x":
			targetCfg = "target/stm32h7x.cfg"
		}
	}

	args := []string{
		"-c", fmt.Sprintf("gdb_port %d", gdbPort),
		"-c", fmt.Sprintf("telnet_port %d", telnetPort),
		"-c", "tcl_port disabled",
		"-f", interfaceCfg,
		"-f", targetCfg,
	}

	cmd := exec.CommandContext(ctx, o.binaryPath, args...)
	if consoleOut != nil {
		cmd.Stdout = consoleOut
		cmd.Stderr = consoleOut
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start OpenOCD: %w", err)
	}

	o.sessions[session.ID] = &openocdSession{
		cmd:        cmd,
		gdbPort:    gdbPort,
		telnetPort: telnetPort,
	}

	session.GDBPort = gdbPort
	session.MonitorPort = telnetPort
	session.PID = cmd.Process.Pid

	return nil
}

func (o *OpenOCDAdapter) StopSession(ctx context.Context, sessionID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	sess, ok := o.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if sess.cmd != nil && sess.cmd.Process != nil {
		if err := sess.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill OpenOCD process: %w", err)
		}
	}

	delete(o.sessions, sessionID)
	return nil
}

func (o *OpenOCDAdapter) ResetSession(ctx context.Context, sessionID string) error {
	return o.sendTelnetCommand(sessionID, "reset halt")
}

func (o *OpenOCDAdapter) LoadProgram(ctx context.Context, sessionID string, programPath string) error {
	// OpenOCD command to flash program
	// program <filename> [verify] [reset] [exit]
	cmd := fmt.Sprintf("program %s verify reset", programPath)
	return o.sendTelnetCommand(sessionID, cmd)
}

func (o *OpenOCDAdapter) ExecuteProgram(ctx context.Context, sessionID string) error {
	return o.sendTelnetCommand(sessionID, "resume")
}

func (o *OpenOCDAdapter) PauseExecution(ctx context.Context, sessionID string) error {
	return o.sendTelnetCommand(sessionID, "halt")
}

func (o *OpenOCDAdapter) ResumeExecution(ctx context.Context, sessionID string) error {
	return o.sendTelnetCommand(sessionID, "resume")
}

func (o *OpenOCDAdapter) GetGDBPort(sessionID string) (int, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	sess, ok := o.sessions[sessionID]
	if !ok {
		return 0, fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.gdbPort, nil
}

func (o *OpenOCDAdapter) CreateSnapshot(ctx context.Context, sessionID string, snapshotPath string) error {
	return fmt.Errorf("snapshots not supported by OpenOCD backend")
}

func (o *OpenOCDAdapter) RestoreSnapshot(ctx context.Context, sessionID string, snapshotPath string) error {
	return fmt.Errorf("snapshots not supported by OpenOCD backend")
}

func (o *OpenOCDAdapter) GetConsoleOutput(ctx context.Context, sessionID string) (string, error) {
	return "", nil
}

func (o *OpenOCDAdapter) sendTelnetCommand(sessionID string, cmd string) error {
	o.mu.RLock()
	sess, ok := o.sessions[sessionID]
	o.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", sess.telnetPort), 2*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to telnet: %w", err)
	}
	defer conn.Close()

	// Read initial banner
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	conn.Read(buf)

	// Send command
	_, err = conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return err
	}

	// Read response (optional, for error checking)
	return nil
}

func (o *OpenOCDAdapter) Step(ctx context.Context, sessionID string, steps int) error {
	return o.sendTelnetCommand(sessionID, "step")
}

func (o *OpenOCDAdapter) StartCoverage(ctx context.Context, sessionID string, outputPath string) error {
	// Enable semihosting for OpenOCD to allow firmware to write coverage data
	if err := o.sendTelnetCommand(sessionID, "arm semihosting enable"); err != nil {
		// Try generic command if arm specific fails, or ignore if already enabled
		// Some targets might use different commands
	}
	return nil
}

func (o *OpenOCDAdapter) StopCoverage(ctx context.Context, sessionID string) error {
	return nil
}

func (o *OpenOCDAdapter) RunForTime(ctx context.Context, sessionID string, duration time.Duration) error {
	// OpenOCD doesn't support run for time.
	// Fallback to Soft Real-time approximation.
	if err := o.ResumeExecution(ctx, sessionID); err != nil {
		return err
	}
	
	select {
	case <-time.After(duration):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	return o.PauseExecution(ctx, sessionID)
}

func (o *OpenOCDAdapter) InjectEvent(ctx context.Context, sessionID string, eventType string, data map[string]interface{}) error {
	// OpenOCD event injection is limited, mostly debug events.
	// Could support 'reset', 'halt' via this generic interface too.
	return fmt.Errorf("event injection not supported for OpenOCD backend")
}
