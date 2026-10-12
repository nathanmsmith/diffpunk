package git

import (
	"context"
	"os/exec"

	"diff-punk/cli/internal/git/private"
)

// Command creates a Git command with its pager disabled.
func Command(ctx context.Context, args ...string) *exec.Cmd {
	return private.Command(ctx, args...)
}
