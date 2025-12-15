package model

import (
	"testing"
	"time"
)

func TestSessionStates(t *testing.T) {
	tests := []struct {
		name  string
		state SessionState
	}{
		{"Created", SessionStateCreated},
		{"Running", SessionStateRunning},
		{"Paused", SessionStatePaused},
		{"Stopped", SessionStateStopped},
		{"Error", SessionStateError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				ID:        "test-id",
				State:     tt.state,
				CreatedAt: time.Now(),
			}

			if session.State != tt.state {
				t.Errorf("Expected state %s, got %s", tt.state, session.State)
			}
		})
	}
}

func TestBackendTypes(t *testing.T) {
	tests := []struct {
		name    string
		backend BackendType
	}{
		{"QEMU", BackendQEMU},
		{"Renode", BackendRenode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				ID:      "test-id",
				Backend: tt.backend,
			}

			if session.Backend != tt.backend {
				t.Errorf("Expected backend %s, got %s", tt.backend, session.Backend)
			}
		})
	}
}

func TestBoardConfig(t *testing.T) {
	config := &BoardConfig{
		Processor: ProcessorConfig{
			Model:     "cortex-m3",
			Frequency: 72000000,
		},
		Memory: MemoryConfig{
			Flash: MemoryRegion{
				Base: 0x08000000,
				Size: 131072,
			},
			RAM: MemoryRegion{
				Base: 0x20000000,
				Size: 20480,
			},
		},
		Peripherals: []PeripheralConfig{
			{
				Type:    "uart",
				Name:    "UART1",
				Address: 0x40013800,
				IRQ:     37,
			},
		},
	}

	if config.Processor.Model != "cortex-m3" {
		t.Errorf("Expected processor cortex-m3, got %s", config.Processor.Model)
	}

	if config.Memory.Flash.Size != 131072 {
		t.Errorf("Expected flash size 131072, got %d", config.Memory.Flash.Size)
	}

	if len(config.Peripherals) != 1 {
		t.Errorf("Expected 1 peripheral, got %d", len(config.Peripherals))
	}
}

func TestCapability(t *testing.T) {
	cap := &Capability{
		Backend: BackendQEMU,
		Processors: []string{
			"cortex-m3",
			"cortex-m4",
		},
		Peripherals: []string{
			"uart",
			"gpio",
		},
		BusTypes: []string{
			"ahb",
			"apb",
		},
		Features: []string{
			"gdb-server",
			"snapshot",
		},
	}

	if cap.Backend != BackendQEMU {
		t.Errorf("Expected backend QEMU, got %s", cap.Backend)
	}

	if len(cap.Processors) != 2 {
		t.Errorf("Expected 2 processors, got %d", len(cap.Processors))
	}

	if len(cap.Features) != 2 {
		t.Errorf("Expected 2 features, got %d", len(cap.Features))
	}
}
