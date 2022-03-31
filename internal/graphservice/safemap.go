package graphservice

import "sync"

type WaitingMap struct {
	v  map[string]chan []string
	mu sync.RWMutex
}

func NewWaitingMap() *WaitingMap {
	return &WaitingMap{
		v: make(map[string]chan []string),
	}
}

func (m *WaitingMap) Get(key string) (chan []string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, exists := m.v[key]
	return value, exists
}

func (m *WaitingMap) Put(key string, value chan []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.v[key] = value
}

func (m *WaitingMap) Delete(key string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	delete(m.v, key)
}
