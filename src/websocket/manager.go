package websocket

import (
	"sync"
)

type Manager struct {
	rooms map[int64]*Room
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[int64]*Room),
	}
}

func (m *Manager) CreateRoom(id int64, name string, problemID int64, maxParticipants int, timeLimit int, creatorID int64) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.rooms[id]; !exists {
		m.rooms[id] = NewRoom(id, name, problemID, maxParticipants, timeLimit, creatorID)
	}
	return m.rooms[id]
}

func (m *Manager) GetRoom(id int64) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	room, exists := m.rooms[id]
	return room, exists
}

func (m *Manager) DeleteRoom(id int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}
