package server

import (
	"diff-punk/cli/internal/git"
	"encoding/json"
)

func (s *snapshotStore) encode() {
	s.data, _ = json.Marshal(struct {
		*git.Snapshot
		Live         bool   `json:"live"`
		RefreshError string `json:"refreshError,omitempty"`
	}{s.snapshot, s.live, s.refreshError})
}
