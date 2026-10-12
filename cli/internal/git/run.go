package git

import (
	"context"
	"io"

	"diff-punk/cli/internal/git/private"
)

// Run runs git diff and builds a snapshot with full-file context.
// Git diagnostics are written to diagnostics.
// Unsupported patch formats are returned as raw output; an empty diff has no snapshot.
func Run(ctx context.Context, args []string, diagnostics io.Writer) (*Snapshot, string, error) {
	return private.Run(ctx, args, diagnostics)
}
