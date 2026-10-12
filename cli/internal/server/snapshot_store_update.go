package server

import (
	"bytes"
	"encoding/json"

	"diff-punk/cli/internal/git"
)

func (s *snapshotStore) update(snapshot *git.Snapshot, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	message := ""
	if err != nil {
		message = err.Error()
	}
	changed := message != s.refreshError
	s.refreshError = message
	if err == nil {
		files, _ := json.Marshal(snapshot.Files)
		if !bytes.Equal(files, s.files) {
			s.snapshot, s.files = snapshot, files
			changed = true
		}
	}
	if !changed {
		return
	}
	s.encode()
	for ch := range s.subscribers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
