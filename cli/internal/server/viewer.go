package server

import (
	"context"
	"sync"
	"time"
)

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
