package git

import "context"

// IsGitRepo reports whether Git can locate a repository from the current directory.
func IsGitRepo(ctx context.Context) bool {
	return Command(ctx, "rev-parse", "--git-dir").Run() == nil
}
