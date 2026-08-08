package configownership

// Snapshot returns the registry contents at one instant.
func (s *Store) Snapshot() map[string][]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]byte, len(s.data))
	for name, payload := range s.data {
		result[name] = payload
	}
	return result
}
