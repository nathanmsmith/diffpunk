package server

import (
	"time"
)

func (s *viewerSession) disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count--
	if s.count == 0 && s.ctx.Err() == nil {
		grace := viewerCloseGrace
		if s.snapshots != nil && s.snapshots.live {
			// EventSource retries after one second. Allow live viewers to
			// reconnect before treating a disconnect as the last tab closing.
			grace = 5 * time.Second
		}
		// Check the timer identity too: a stopped timer's callback might already
		// be waiting for the lock when a new connection arrives.
		var timer *time.Timer
		timer = time.AfterFunc(grace, func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			if s.count == 0 && s.timer == timer {
				s.stop()
			}
		})
		s.timer = timer
	}
}
