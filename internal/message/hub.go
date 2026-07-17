package message

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

// Client represents a single WebSocket connection.
type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan []byte
	hub    *Hub
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// We currently ignore client-sent messages; all actions go through REST.
	}
}

// Hub manages a set of WebSocket clients keyed by user ID.
type Hub struct {
	// clients maps userID to all active connections for that user.
	clients    map[uint][]*Client
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint][]*Client),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
	}
}

// Run starts the hub's event loop. Should be called as a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = append(h.clients[client.UserID], client)
			h.mu.Unlock()
			log.Printf("[ws] user %d connected (total conns for user: %d)", client.UserID, len(h.clients[client.UserID]))

		case client := <-h.unregister:
			h.mu.Lock()
			conns := h.clients[client.UserID]
			filtered := make([]*Client, 0, len(conns))
			for _, c := range conns {
				if c != client {
					filtered = append(filtered, c)
				}
			}
			if len(filtered) == 0 {
				delete(h.clients, client.UserID)
			} else {
				h.clients[client.UserID] = filtered
			}
			h.mu.Unlock()
			close(client.Send)
			log.Printf("[ws] user %d disconnected", client.UserID)
		}
	}
}

// SendToUser sends a JSON message to all connections of a given user.
func (h *Hub) SendToUser(userID uint, data []byte) {
	h.mu.RLock()
	conns := h.clients[userID]
	h.mu.RUnlock()

	for _, c := range conns {
		select {
		case c.Send <- data:
		default:
			// Client's send buffer is full; skip this connection.
			log.Printf("[ws] send buffer full for user %d, dropping message", userID)
		}
	}
}

// ServeWS handles a WebSocket request: upgrades the connection, creates a Client,
// registers it with the Hub, and starts read/write pumps. Blocks until the
// connection closes.
func (h *Hub) ServeWS(conn *websocket.Conn, userID uint) {
	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 64),
		hub:    h,
	}

	h.register <- client

	go client.writePump()
	client.readPump()
}
