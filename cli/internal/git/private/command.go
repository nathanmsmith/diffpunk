package private

import (
	"context"
	"os/exec"
)

// Command creates a Git command with its pager disabled.
func Command(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "git", append([]string{"--no-pager"}, args...)...)
}
