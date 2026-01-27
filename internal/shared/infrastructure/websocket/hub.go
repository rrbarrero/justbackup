package websocket

import (
	"log"
	"sync"

	"golang.org/x/net/websocket"
)

type Hub struct {
	// Registered clients.
	clients map[*websocket.Conn]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *websocket.Conn

	// Unregister requests from clients.
	unregister chan *websocket.Conn

	mu sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		clients:    make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()
			log.Println("New WebSocket client connected")
		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
				log.Println("WebSocket client disconnected")
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for conn := range h.clients {
				// Write message to client
				if _, err := conn.Write(message); err != nil {
					log.Printf("Error writing to websocket: %v", err)
					conn.Close()
					delete(h.clients, conn)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcast <- message:
	default:
		// Drop message if hub is not running or broadcast channel is full
	}
}

func (h *Hub) Handle(ws *websocket.Conn) {
	h.register <- ws
	defer func() {
		h.unregister <- ws
	}()

	// Keep connection alive and read (even if we don't expect messages from client for now)
	// This loop is necessary to detect disconnection
	buf := make([]byte, 1024)
	for {
		_, err := ws.Read(buf)
		if err != nil {
			break
		}
	}
}
