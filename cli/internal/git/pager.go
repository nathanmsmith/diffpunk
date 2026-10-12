package git

import (
	"context"
	"strings"
	"time"

	"diff-punk/cli/internal/git/private"
)

// Pager builds a snapshot from supplied Git output without rerunning the diff.
// Unsupported patches return an error so the CLI can print the original bytes.
func Pager(ctx context.Context, raw []byte) (*Snapshot, error) {
	files, err := private.ParsePagerFiles(string(raw))
	if err != nil {
		return nil, err
	}
	for i := range files {
		private.ReconstructContext(ctx, &files[i])
		files[i].Patch = private.NormalizePagerPrefixes(files[i].Patch)
		files[i].FullPatch = private.NormalizePagerPrefixes(files[i].FullPatch)
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	root, _ := Command(ctx, "rev-parse", "--show-toplevel").Output()
	snapshot := &Snapshot{
		Repository: strings.TrimSpace(string(root)), Arguments: []string{},
		CreatedAt: time.Now().UTC().Format(time.RFC3339), Files: private.SortFiles(files),
	}
	return snapshot, nil
}
