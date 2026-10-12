package server

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"time"

	"diff-punk/cli/internal/git"
)

// Refresh reruns the original comparison. Only one capture runs at a time.
type Refresh func(context.Context) (*git.Snapshot, error)

type snapshotStore struct {
	mu           sync.Mutex
	snapshot     *git.Snapshot
	files        []byte
	data         []byte
	live         bool
	refreshError string
	subscribers  map[chan struct{}]struct{}
}

func newSnapshotStore(snapshot *git.Snapshot, live bool) *snapshotStore {
	s := &snapshotStore{snapshot: snapshot, live: live, subscribers: make(map[chan struct{}]struct{})}
	s.files, _ = json.Marshal(snapshot.Files)
	s.encode()
	return s
}

func (s *snapshotStore) encode() {
	s.data, _ = json.Marshal(struct {
		*git.Snapshot
		Live         bool   `json:"live"`
		RefreshError string `json:"refreshError,omitempty"`
	}{s.snapshot, s.live, s.refreshError})
}

func (s *snapshotStore) read() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

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

func (s *snapshotStore) watch(ctx context.Context, refresh Refresh) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot, err := refresh(ctx)
			if ctx.Err() != nil {
				return
			}
			s.update(snapshot, err)
		}
	}
}
