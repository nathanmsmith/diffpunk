package server

import (
	"encoding/json"

	"diff-punk/cli/internal/git"
)

func (s *snapshotStore) encode() {
	s.data, _ = json.Marshal(struct {
		*git.Snapshot
		Live         bool   `json:"live"`
		RefreshError string `json:"refreshError,omitempty"`
	}{s.snapshot, s.live, s.refreshError})
}
