package websocket

import (
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
)

type Manager struct {
	rooms map[pgtype.Int8]*Room
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[pgtype.Int8]*Room),
	}
}

func (m *Manager) CreateRoom(id int64, name string, problemID pgtype.Int8, maxParticipants int, timeLimit int, creatorID pgtype.Int8) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	pgID := pgtype.Int8{Int64: id, Valid: true}
	if _, exists := m.rooms[pgID]; !exists {
		m.rooms[pgID] = NewRoom(pgID, name, problemID, maxParticipants, timeLimit, creatorID)
	}
	return m.rooms[pgID]
}

func (m *Manager) GetRoom(id pgtype.Int8) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	room, exists := m.rooms[id]
	return room, exists
}

func (m *Manager) DeleteRoom(id pgtype.Int8) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}
