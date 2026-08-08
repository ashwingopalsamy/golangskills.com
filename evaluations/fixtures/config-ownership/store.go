package configownership

import "sync"

// Store is a concurrent registry for opaque configuration payloads.
type Store struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewStore() *Store {
	return &Store{data: make(map[string][]byte)}
}

func (s *Store) Put(name string, payload []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[name] = payload
}

func (s *Store) Get(name string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	payload, ok := s.data[name]
	return payload, ok
}
