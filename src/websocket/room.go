package websocket

import (
	"context"
	"sync"

	"github.com/coder/websocket"
)

type Room struct {
	ID      string
	clients map[string]*websocket.Conn // key is account ID, value is WebSocket connection
	mu      sync.Mutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID:      id,
		clients: make(map[string]*websocket.Conn),
	}
}

func (r *Room) AddClient(accountID string, conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[accountID] = conn
}

func (r *Room) RemoveClient(accountID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, accountID)
}

func (r *Room) Broadcast(message []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, conn := range r.clients {
		conn.Write(context.Background(), websocket.MessageText, message)
	}
}

func (r *Room) GetClients() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	clients := make([]string, 0, len(r.clients))
	for accountID := range r.clients {
		clients = append(clients, accountID)
	}
	return clients
}
