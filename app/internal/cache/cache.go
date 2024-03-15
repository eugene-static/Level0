package cache

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type Manager struct {
	*sync.RWMutex
	Storage map[string][]byte
}

func New(mu *sync.RWMutex) *Manager {
	return &Manager{
		RWMutex: mu,
		Storage: make(map[string][]byte),
	}
}

func (m *Manager) Set(key string, data []byte) {
	m.Lock()
	defer m.Unlock()
	m.Storage[key] = data
}

func (m *Manager) Get(key string) ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	data, found := m.Storage[key]
	if !found {
		return nil, ErrNotFound
	}
	return data, nil
}
