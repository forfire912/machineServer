package gdb

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

// Server represents a GDB server that bridges to backend emulators
type Server struct {
	listener net.Listener
	port     int
	sessions map[string]*Session
	mu       sync.RWMutex
}

// Session represents a GDB debugging session
type Session struct {
	SessionID  string
	BackendPort int
	conn       net.Conn
	backendConn net.Conn
}

// NewServer creates a new GDB server
func NewServer(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to start GDB server: %w", err)
	}

	return &Server{
		listener: listener,
		port:     port,
		sessions: make(map[string]*Session),
	}, nil
}

// Start starts the GDB server
func (s *Server) Start() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		// Read GDB RSP packet
		packet, err := s.readPacket(reader)
		if err != nil {
			return
		}

		// Process packet and send response
		response := s.processPacket(packet)
		if err := s.sendPacket(conn, response); err != nil {
			return
		}
	}
}

func (s *Server) readPacket(reader *bufio.Reader) (string, error) {
	// GDB RSP packet format: $<data>#<checksum>
	// This is a simplified implementation
	data, err := reader.ReadString('#')
	if err != nil {
		return "", err
	}

	// Read checksum (2 bytes)
	checksum := make([]byte, 2)
	if _, err := reader.Read(checksum); err != nil {
		return "", err
	}

	return data, nil
}

func (s *Server) sendPacket(conn net.Conn, data string) error {
	// Calculate checksum
	checksum := 0
	for _, b := range data {
		checksum += int(b)
	}
	checksum = checksum % 256

	// Send packet
	packet := fmt.Sprintf("$%s#%02x", data, checksum)
	_, err := conn.Write([]byte(packet))
	return err
}

func (s *Server) processPacket(packet string) string {
	// Handle common GDB commands
	// This is a simplified implementation
	if len(packet) == 0 {
		return ""
	}

	cmd := packet[1] // Skip '$'

	switch cmd {
	case 'q':
		return "OK"
	case 'g':
		return "00000000"
	case 'c':
		return "S05"
	default:
		return ""
	}
}

// Stop stops the GDB server
func (s *Server) Stop() error {
	return s.listener.Close()
}

// AttachSession attaches a GDB session to a simulation session
func (s *Server) AttachSession(sessionID string, backendPort int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &Session{
		SessionID:   sessionID,
		BackendPort: backendPort,
	}

	s.sessions[sessionID] = session
	return nil
}

// DetachSession detaches a GDB session
func (s *Server) DetachSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.sessions[sessionID]; ok {
		if session.conn != nil {
			session.conn.Close()
		}
		if session.backendConn != nil {
			session.backendConn.Close()
		}
		delete(s.sessions, sessionID)
	}

	return nil
}
