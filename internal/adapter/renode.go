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
			// ARM Cortex-M
			"cortex-m0", "cortex-m0+", "cortex-m3", "cortex-m4", "cortex-m7", "cortex-m23", "cortex-m33",
			// ARM Cortex-A
			"cortex-a7", "cortex-a9", "cortex-a53", "cortex-a72",
			// ARM Cortex-R
			"cortex-r5", "cortex-r52",
			// RISC-V
			"riscv32", "riscv64", "vexriscv", "rocket", "ariane", "ibex",
			// Other
			"sparc", "ppc", "xtensa", "x86",
		},
		Peripherals: []string{
			"uart", "usart", "lpuart",
			"gpio",
			"spi", "qspi",
			"i2c",
			"timer", "rtc", "watchdog",
			"adc", "dac",
			"can", "fdcan",
			"ethernet", "gem", "macb",
			"usb", "usb-otg",
			"sd-card", "sdmmc",
			"display", "ltdc",
			"radio", "nrf-radio", "ieee802.15.4",
			"sensor", "imu", "temp-sensor", "humidity-sensor",
			"crypto", "rng", "aes",
		},
		BusTypes: []string{
			"ahb", "apb", "axi",
			"wishbone",
			"pci",
			"i2c", "spi", "uart",
		},
		Features: []string{
			"gdb-server",
			"snapshot",
			"monitor",
			"multi-node",
			"time-sync",
			"robot-framework",
			"wireshark-logging",
		},
		Boards: []string{
			// ST
			"stm32f4_discovery", "stm32f746g_disco", "stm32f072b_disco", "nucleo_f103rb", "nucleo_l476rg",
			// Nordic
			"nrf52840dk", "nrf52dk", "microbit",
			// SiFive
			"hifive1", "hifive1_revb", "hifive_unleashed",
			// Microchip
			"sam_e70_xplained", "polarfire_soc",
			// NXP
			"imxrt1064_evk", "k64f",
			// Other
			"arduino_uno", "zedboard", "pico",
		},
	}, nil
}

func (r *RenodeAdapter) StartSession(ctx context.Context, session *model.Session, config *model.BoardConfig, consoleOut io.Writer) error {
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

	// Generate .resc file
	rescContent := fmt.Sprintf(":name: %s\n", session.Name)
	
	if config != nil && config.Board != "" {
		// Use standard board script
		// Renode scripts are typically in @scripts/boards/
		// We assume the board name matches the script name or we need a mapping
		rescContent += fmt.Sprintf("include @scripts/boards/%s.resc\n", config.Board)
	} else if config != nil && config.Processor.Model != "" {
		// Create custom platform
		rescContent += "mach create\n"
		rescContent += fmt.Sprintf("machine LoadPlatformDescriptionFromString \"cpu: CPU.%s @ sysbus\"\n", config.Processor.Model)
		
		// Add memory
		if config.Memory.RAM.Size > 0 {
			rescContent += fmt.Sprintf("sysbus: { 0x%x : Memory.MappedMemory @ sysbus 0x%x }\n", config.Memory.RAM.Base, config.Memory.RAM.Size)
		}
	} else {
		// Fallback default
		rescContent += "include @scripts/boards/stm32f4_discovery.resc\n"
	}

	// Configure GDB
	rescContent += "machine StartGdbServer 3333\n" // Internal port, we map it via telnet later or use direct port if possible
	// Note: Renode's StartGdbServer binds to a port. We need to make sure it matches what we expect or use dynamic port.
	// Renode command: machine StartGdbServer <port>
	rescContent += fmt.Sprintf("machine StartGdbServer %d\n", gdbPort)

	// Save .resc file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("session_%s.resc", session.ID))
	if err := os.WriteFile(tmpFile, []byte(rescContent), 0644); err != nil {
		return fmt.Errorf("failed to create resc file: %w", err)
	}

	// Build Renode command
	args := []string{
		"--disable-gui",
		"--port", fmt.Sprintf("%d", telnetPort),
		tmpFile, // Load the generated script
	}

	cmd := exec.CommandContext(ctx, r.binaryPath, args...)
	if consoleOut != nil {
		cmd.Stdout = consoleOut
		cmd.Stderr = consoleOut
	}

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

func (r *RenodeAdapter) sendTelnetCommand(sessionID string, cmd string) error {
	r.mu.RLock()
	sess, ok := r.sessions[sessionID]
	r.mu.RUnlock()

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

func (r *RenodeAdapter) ResetSession(ctx context.Context, sessionID string) error {
	return r.sendTelnetCommand(sessionID, "machine Reset")
}

func (r *RenodeAdapter) LoadProgram(ctx context.Context, sessionID string, programPath string) error {
	// Load program via Renode monitor commands
	// sysbus LoadELF @path
	cmd := fmt.Sprintf("sysbus LoadELF @%s", programPath)
	return r.sendTelnetCommand(sessionID, cmd)
}

func (r *RenodeAdapter) ExecuteProgram(ctx context.Context, sessionID string) error {
	return r.sendTelnetCommand(sessionID, "start")
}

func (r *RenodeAdapter) PauseExecution(ctx context.Context, sessionID string) error {
	return r.sendTelnetCommand(sessionID, "pause")
}

func (r *RenodeAdapter) ResumeExecution(ctx context.Context, sessionID string) error {
	return r.sendTelnetCommand(sessionID, "start")
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
	cmd := fmt.Sprintf("save @%s", snapshotPath)
	return r.sendTelnetCommand(sessionID, cmd)
}

func (r *RenodeAdapter) RestoreSnapshot(ctx context.Context, sessionID string, snapshotPath string) error {
	cmd := fmt.Sprintf("load @%s", snapshotPath)
	return r.sendTelnetCommand(sessionID, cmd)
}

func (r *RenodeAdapter) GetConsoleOutput(ctx context.Context, sessionID string) (string, error) {
	return "", nil
}

func (r *RenodeAdapter) Step(ctx context.Context, sessionID string, steps int) error {
	// Renode supports 'step' command
	// Note: Renode step command might be blocking or async depending on config
	// step <count>
	cmd := fmt.Sprintf("step %d", steps)
	return r.sendTelnetCommand(sessionID, cmd)
}

func (r *RenodeAdapter) StartCoverage(ctx context.Context, sessionID string, outputPath string) error {
	// Renode command: cpu LogCoverage @path
	// We assume 'cpu' is a valid alias for the main processor
	cmd := fmt.Sprintf("cpu LogCoverage @%s", outputPath)
	return r.sendTelnetCommand(sessionID, cmd)
}

func (r *RenodeAdapter) StopCoverage(ctx context.Context, sessionID string) error {
	return r.sendTelnetCommand(sessionID, "cpu DisableLogCoverage")
}

func (r *RenodeAdapter) RunForTime(ctx context.Context, sessionID string, duration time.Duration) error {
	// Renode supports 'machine Advance <time>'
	// Format: 10ms, 1s, etc.
	cmd := fmt.Sprintf("machine Advance %s", duration.String())
	return r.sendTelnetCommand(sessionID, cmd)
}

func (r *RenodeAdapter) InjectEvent(ctx context.Context, sessionID string, eventType string, data map[string]interface{}) error {
	// Scheme 4: Event Injection
	// Example: GPIO toggle, UART input
	switch eventType {
	case "gpio":
		// data: {"port": "gpioPort", "pin": 0, "state": true}
		port, _ := data["port"].(string)
		pin, _ := data["pin"].(int)
		state, _ := data["state"].(bool)
		cmd := fmt.Sprintf("%s.%d SetState %t", port, pin, state)
		return r.sendTelnetCommand(sessionID, cmd)
	case "uart":
		// data: {"uart": "sysbus.uart0", "text": "hello"}
		uart, _ := data["uart"].(string)
		text, _ := data["text"].(string)
		cmd := fmt.Sprintf("%s WriteString \"%s\"", uart, text)
		return r.sendTelnetCommand(sessionID, cmd)
	default:
		return fmt.Errorf("unsupported event type: %s", eventType)
	}
}
