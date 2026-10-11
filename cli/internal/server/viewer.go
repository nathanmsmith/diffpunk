package server

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

const viewerCloseGrace = 500 * time.Millisecond

// Each viewer holds an event stream open. Only the last disconnect starts the
// shutdown timer, and a reconnect cancels it so refreshes can keep the CLI alive.
type viewerSession struct {
	ctx   context.Context
	stop  context.CancelFunc
	mu    sync.Mutex
	count int
	timer *time.Timer
}

func (s *viewerSession) connect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count++
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
}

func (s *viewerSession) disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count--
	if s.count == 0 && s.ctx.Err() == nil {
		// Check the timer identity too: a stopped timer's callback might already
		// be waiting for the lock when a new connection arrives.
		var timer *time.Timer
		timer = time.AfterFunc(viewerCloseGrace, func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			if s.count == 0 && s.timer == timer {
				s.stop()
			}
		})
		s.timer = timer
	}
}

func (s *viewerSession) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	s.connect()
	defer s.disconnect()
	if _, err := io.WriteString(w, "retry: 1000\n\n"); err != nil {
		return
	}
	flusher.Flush()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
