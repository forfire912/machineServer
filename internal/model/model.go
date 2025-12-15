package model

import (
	"time"
)

// BackendType represents the simulation backend type
type BackendType string

const (
	BackendQEMU    BackendType = "qemu"
	BackendRenode  BackendType = "renode"
	BackendOpenOCD BackendType = "openocd"
)

// SessionState represents the state of a simulation session
type SessionState string

const (
	SessionStateCreated  SessionState = "created"
	SessionStateRunning  SessionState = "running"
	SessionStatePaused   SessionState = "paused"
	SessionStateStopped  SessionState = "stopped"
	SessionStateError    SessionState = "error"
)

// Session represents a simulation session
type Session struct {
	ID          string       `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name"`
	Backend     BackendType  `json:"backend"`
	BoardConfig string       `json:"board_config"` // JSON string
	State       SessionState `json:"state"`
	GDBPort     int          `json:"gdb_port,omitempty"`
	MonitorPort int          `json:"monitor_port,omitempty"`
	PID         int          `json:"pid,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	UserID      string       `json:"user_id,omitempty"`
}

// BoardConfig represents hardware configuration
type BoardConfig struct {
	Board       string                 `json:"board,omitempty"` // Pre-defined board name
	Processor   ProcessorConfig        `json:"processor"`
	Memory      MemoryConfig           `json:"memory"`
	Peripherals []PeripheralConfig     `json:"peripherals,omitempty"`
}

// ProcessorConfig represents processor configuration
type ProcessorConfig struct {
	Model     string `json:"model"`      // e.g., "cortex-m3", "cortex-m4", "riscv32"
	Frequency int    `json:"frequency"`  // MHz
}

// MemoryConfig represents memory layout
type MemoryConfig struct {
	Flash MemoryRegion `json:"flash"`
	RAM   MemoryRegion `json:"ram"`
}

// MemoryRegion represents a memory region
type MemoryRegion struct {
	Base uint64 `json:"base"`
	Size uint64 `json:"size"`
}

// PeripheralConfig represents peripheral configuration
type PeripheralConfig struct {
	Type    string `json:"type"`    // e.g., "uart", "gpio", "spi"
	Name    string `json:"name"`
	Address uint64 `json:"address"`
	IRQ     int    `json:"irq,omitempty"`
}

// Program represents an uploaded program
type Program struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Format    string    `json:"format"` // elf, binary, hex
	Size      int64     `json:"size"`
	Path      string    `json:"path"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
	UserID    string    `json:"user_id,omitempty"`
}

// Snapshot represents a simulation snapshot
type Snapshot struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	SessionID   string    `json:"session_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Size        int64     `json:"size"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
}

// Job represents an async job
type Job struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	Type       string    `json:"type"` // test, coverage, trace
	Status     string    `json:"status"`
	Result     string    `json:"result,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CoSimSession represents a co-simulation session
type CoSimSession struct {
	ID         string           `json:"id" gorm:"primaryKey"`
	Status     string           `json:"status"`
	CreatedAt  time.Time        `json:"created_at"`
	StartedAt  *time.Time       `json:"started_at,omitempty"`
	SyncCount  int64            `json:"sync_count"`
	TimeNS     int64            `json:"time_ns"`
	Components []CoSimComponent `json:"components" gorm:"foreignKey:CoSimID"`
}

// CoSimComponent represents a component in co-simulation
type CoSimComponent struct {
	ID        string `json:"id" gorm:"primaryKey"`
	CoSimID   string `json:"cosim_id"`
	Type      string `json:"type"`
	Config    string `json:"config"` // JSON string
	Status    string `json:"status"`
	SessionID string `json:"session_id,omitempty"` // Linked simulation session ID
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details,omitempty"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
}

// Capability represents backend capabilities
type Capability struct {
	Backend     BackendType `json:"backend"`
	Processors  []string    `json:"processors"`
	Peripherals []string    `json:"peripherals"`
	BusTypes    []string    `json:"bus_types"`
	Features    []string    `json:"features"`
	Boards      []string    `json:"boards"`
}
