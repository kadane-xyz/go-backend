package websocket

import (
	"sync"
)

type Manager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

func (m *Manager) CreateRoom(id string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.rooms[id]; !exists {
		m.rooms[id] = NewRoom(id)
	}
	return m.rooms[id]
}

func (m *Manager) GetRoom(id string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	room, exists := m.rooms[id]
	return room, exists
}

func (m *Manager) DeleteRoom(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}

