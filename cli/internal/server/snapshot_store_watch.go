package server

import (
	"context"
	"time"
)

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
