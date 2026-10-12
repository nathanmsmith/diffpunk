package server

import (
	"diff-punk/cli/internal/git"
	"encoding/json"
)

func newSnapshotStore(snapshot *git.Snapshot, live bool) *snapshotStore {
	s := &snapshotStore{snapshot: snapshot, live: live, subscribers: make(map[chan struct{}]struct{})}
	s.files, _ = json.Marshal(snapshot.Files)
	s.encode()
	return s
}
