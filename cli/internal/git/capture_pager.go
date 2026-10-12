package git

import (
	"context"

	"diff-punk/cli/internal/git/private"
)

// CapturePager builds a snapshot from supplied Git output without rerunning the diff.
// Unsupported patches return an error so the CLI can print the original bytes.
func CapturePager(ctx context.Context, raw []byte) (*Snapshot, error) {
	return private.CapturePager(ctx, raw)
}
