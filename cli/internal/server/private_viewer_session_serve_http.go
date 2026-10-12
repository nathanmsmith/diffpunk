package server

import (
	"io"
	"net/http"
	"time"
)

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
	var updates <-chan struct{}
	if s.snapshots != nil {
		ch, unsubscribe := s.snapshots.subscribe()
		defer unsubscribe()
		updates = ch
	}
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-r.Context().Done():
			return
		case <-updates:
			if _, err := io.WriteString(w, "event: snapshot\ndata: changed\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
