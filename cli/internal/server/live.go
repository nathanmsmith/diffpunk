package server

import (
	"context"
	"diff-punk/cli/internal/git"
	"sync"
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
