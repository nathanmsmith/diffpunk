package server

func (s *snapshotStore) read() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}
