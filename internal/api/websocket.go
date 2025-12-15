package api

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// StreamHub manages WebSocket connections
type StreamHub struct {
	clients    map[string]map[*websocket.Conn]bool
	broadcast  chan StreamMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// Client represents a WebSocket client
type Client struct {
	SessionID string
	Conn      *websocket.Conn
	Send      chan []byte
}

// StreamMessage represents a message to broadcast
type StreamMessage struct {
	SessionID string
	Type      string
	Data      []byte
}

// NewStreamHub creates a new stream hub
func NewStreamHub() *StreamHub {
	return &StreamHub{
		clients:    make(map[string]map[*websocket.Conn]bool),
		broadcast:  make(chan StreamMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub
func (h *StreamHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.SessionID] == nil {
				h.clients[client.SessionID] = make(map[*websocket.Conn]bool)
			}
			h.clients[client.SessionID][client.Conn] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.SessionID]; ok {
				if _, ok := clients[client.Conn]; ok {
					delete(clients, client.Conn)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.SessionID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[message.SessionID]; ok {
				for conn := range clients {
					select {
					case client := <-h.getClientBySess(message.SessionID, conn):
						if client != nil {
							select {
							case client.Send <- message.Data:
							default:
								close(client.Send)
								delete(clients, conn)
							}
						}
					default:
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *StreamHub) getClientBySess(sessionID string, conn *websocket.Conn) chan *Client {
	ch := make(chan *Client, 1)
	go func() {
		h.mu.RLock()
		defer h.mu.RUnlock()
		if clients, ok := h.clients[sessionID]; ok {
			if _, exists := clients[conn]; exists {
				ch <- &Client{SessionID: sessionID, Conn: conn}
				return
			}
		}
		ch <- nil
	}()
	return ch
}

// BroadcastToSession sends a message to all clients of a session
func (h *StreamHub) BroadcastToSession(sessionID string, msgType string, data []byte) {
	h.broadcast <- StreamMessage{
		SessionID: sessionID,
		Type:      msgType,
		Data:      data,
	}
}

// StreamConsole handles console output WebSocket connections
func (h *Handler) StreamConsole(c *gin.Context) {
	sessionID := c.Param("id")

	// Verify session exists
	if _, err := h.service.GetSession(sessionID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		SessionID: sessionID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
	}

	// Register client
	if h.streamHub != nil {
		h.streamHub.register <- client
	}

	// Start goroutines
	go h.writePump(client)
	go h.readPump(client)
}

func (h *Handler) readPump(client *Client) {
	defer func() {
		if h.streamHub != nil {
			h.streamHub.unregister <- client
		}
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

func (h *Handler) writePump(client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
