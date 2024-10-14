package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type Room struct {
	ID              int64
	Name            string
	ProblemID       int64
	CreatedAt       time.Time
	Status          string
	MaxParticipants int
	TimeLimit       int
	CreatorID       int64
	clients         map[int64]*websocket.Conn // key is account ID, value is WebSocket connection
	mu              sync.Mutex
}

func NewRoom(id int64, name string, problemID int64, maxParticipants int, timeLimit int, creatorID int64) *Room {
	return &Room{
		ID:              id,
		Name:            name,
		ProblemID:       problemID,
		CreatedAt:       time.Now(),
		Status:          "open",
		MaxParticipants: maxParticipants,
		TimeLimit:       timeLimit,
		CreatorID:       creatorID,
		clients:         make(map[int64]*websocket.Conn),
	}
}

func (r *Room) AddClient(accountID int64, conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[accountID] = conn
}

func (r *Room) RemoveClient(accountID int64) {
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

func (r *Room) GetClients() []int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	clients := make([]int64, 0, len(r.clients))
	for accountID := range r.clients {
		clients = append(clients, accountID)
	}
	return clients
}
