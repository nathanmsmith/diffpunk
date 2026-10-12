package server

func (s *snapshotStore) subscribe() (chan struct{}, func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := make(chan struct{}, 1)
	s.subscribers[ch] = struct{}{}
	// Fetch the current state on every connection, including reconnects.
	ch <- struct{}{}
	return ch, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.subscribers, ch)
	}
}
