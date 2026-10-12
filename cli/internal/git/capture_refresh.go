package git

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"diff-punk/cli/internal/git/private"
)

// CaptureRefresh returns diagnostics to the viewer instead of repeatedly printing them.
func CaptureRefresh(ctx context.Context, args []string) (*Snapshot, string, error) {
	var diagnostics bytes.Buffer
	snapshot, raw, err := private.Capture(ctx, args, &diagnostics)
	if err != nil && diagnostics.Len() > 0 {
		err = fmt.Errorf("%w: %s", err, strings.TrimSpace(diagnostics.String()))
	}
	return snapshot, raw, err
}
