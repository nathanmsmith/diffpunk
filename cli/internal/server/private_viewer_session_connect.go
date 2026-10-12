package server

func (s *viewerSession) connect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count++
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
}
