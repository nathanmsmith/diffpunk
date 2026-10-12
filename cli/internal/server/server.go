package server

import (
	"context"
	"sync"
	"time"

	"diff-punk/cli/internal/git"
)

const devServerAddress = "http://127.0.0.1:5173"

const devBase = "/__diff_punk_dev/"

const devIdentity = "diff-punk-vite-v1"

// Refresh reruns the original comparison. Only one comparison runs at a time.
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

const viewerCloseGrace = 500 * time.Millisecond

// Each viewer holds an event stream open. Only the last disconnect starts the
// shutdown timer, and a reconnect cancels it so refreshes can keep the CLI alive.
type viewerSession struct {
	ctx       context.Context
	stop      context.CancelFunc
	mu        sync.Mutex
	count     int
	timer     *time.Timer
	snapshots *snapshotStore
}
