package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// loadUntracked uses Git for ignore rules, pathspecs, quoting, binary detection,
// and symlink contents. It never changes the index or writes repository objects.
func loadUntracked(ctx context.Context, root string, paths []string, relativePrefix string, repositoryWide bool) ([]File, error) {
	args := []string{"ls-files", "--others", "--exclude-standard", "--full-name", "-z", "--"}
	if len(paths) == 0 {
		paths = []string{":(top)**"}
	}
	cmd := Command(ctx, append(args, paths...)...)
	if repositoryWide {
		cmd.Dir = root
	}
	raw, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list untracked files: %w", err)
	}
	files := []File{}
	for _, name := range strings.Split(string(raw), "\x00") {
		if name == "" || !strings.HasPrefix(name, relativePrefix) {
			continue
		}
		info, err := os.Lstat(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return nil, fmt.Errorf("inspect untracked file %q: %w", name, err)
		}
		// An untracked embedded repository is listed as a directory by Git.
		if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		name = strings.TrimPrefix(name, relativePrefix)
		cmd := Command(ctx, "-c", "diff.noprefix=false", "-c", "diff.mnemonicPrefix=false", "diff", "--no-index", "--no-color", "--no-ext-diff", "--no-textconv", "--src-prefix=a/", "--dst-prefix=b/", "--", os.DevNull, name)
		cmd.Dir = filepath.Join(root, filepath.FromSlash(relativePrefix))
		patch, err := cmd.Output()
		var exit *exec.ExitError
		if err != nil && (!errors.As(err, &exit) || exit.ExitCode() != 1) {
			return nil, fmt.Errorf("read untracked file %q: %w", name, err)
		}
		parsed, err := parseFiles(string(patch))
		if err != nil || len(parsed) != 1 {
			return nil, fmt.Errorf("invalid patch for untracked file %q", name)
		}
		parsed[0].Kind = "untracked"
		files = append(files, parsed[0])
	}
	return files, nil
}
