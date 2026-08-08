package replicareplay

import "sync"

type record struct {
	request Request
	result  Result
}

// Backend stands in for state shared by independent service handles.
type Backend struct {
	mu      sync.Mutex
	records map[string]record
	entries []Result
	nextID  int64
}

func NewBackend() *Backend {
	return &Backend{records: make(map[string]record)}
}

type Store struct {
	backend *Backend
}

func NewStore(backend *Backend) *Store {
	return &Store{backend: backend}
}

func (s *Store) lookup(key string) (record, bool) {
	s.backend.mu.Lock()
	defer s.backend.mu.Unlock()
	stored, ok := s.backend.records[key]
	return stored, ok
}

func (s *Store) appendEntry(request Request) Result {
	s.backend.mu.Lock()
	defer s.backend.mu.Unlock()
	s.backend.nextID++
	result := Result{EntryID: s.backend.nextID, Account: request.Account, Amount: request.Amount}
	s.backend.entries = append(s.backend.entries, result)
	return result
}

func (s *Store) save(key string, stored record) {
	s.backend.mu.Lock()
	defer s.backend.mu.Unlock()
	s.backend.records[key] = stored
}

func (s *Store) Entries() []Result {
	s.backend.mu.Lock()
	defer s.backend.mu.Unlock()
	return append([]Result(nil), s.backend.entries...)
}
