package server

import (
	"encoding/json"

	"diff-punk/cli/internal/git"
)

func newSnapshotStore(snapshot *git.Snapshot, live bool) *snapshotStore {
	s := &snapshotStore{snapshot: snapshot, live: live, subscribers: make(map[chan struct{}]struct{})}
	s.files, _ = json.Marshal(snapshot.Files)
	s.encode()
	return s
}
