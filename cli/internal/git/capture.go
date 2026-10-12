package git

import (
	"context"
	"os"

	"diff-punk/cli/internal/git/private"
)

// Capture runs git diff and captures its files and full context.
// Unsupported patch formats are returned as raw output; an empty diff has no snapshot.
func Capture(ctx context.Context, args []string) (*Snapshot, string, error) {
	return private.Capture(ctx, args, os.Stderr)
}
