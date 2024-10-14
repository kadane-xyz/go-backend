package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5/pgtype"
)

type Room struct {
	ID              pgtype.Int8
	Name            string
	ProblemID       pgtype.Int8
	CreatedAt       time.Time
	Status          string
	MaxParticipants int
	TimeLimit       int
	CreatorID       pgtype.Int8
	clients         map[pgtype.Int8]*websocket.Conn // key is account ID, value is WebSocket connection
	mu              sync.Mutex
}

func NewRoom(id pgtype.Int8, name string, problemID pgtype.Int8, maxParticipants int, timeLimit int, creatorID pgtype.Int8) *Room {
	return &Room{
		ID:              id,
		Name:            name,
		ProblemID:       problemID,
		CreatedAt:       time.Now(),
		Status:          "open",
		MaxParticipants: maxParticipants,
		TimeLimit:       timeLimit,
		CreatorID:       creatorID,
		clients:         make(map[pgtype.Int8]*websocket.Conn),
	}
}

func (r *Room) AddClient(accountID pgtype.Int8, conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[accountID] = conn
}

func (r *Room) RemoveClient(accountID pgtype.Int8) {
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

func (r *Room) GetClients() []pgtype.Int8 {
	r.mu.Lock()
	defer r.mu.Unlock()
	clients := make([]pgtype.Int8, 0, len(r.clients))
	for accountID := range r.clients {
		clients = append(clients, accountID)
	}
	return clients
}
