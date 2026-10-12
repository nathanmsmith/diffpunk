package git

import (
	"context"

	"diff-punk/cli/internal/git/private"
)

// Pager builds a snapshot from supplied Git output without rerunning the diff.
// Unsupported patches return an error so the CLI can print the original bytes.
func Pager(ctx context.Context, raw []byte) (*Snapshot, error) {
	return private.Pager(ctx, raw)
}
