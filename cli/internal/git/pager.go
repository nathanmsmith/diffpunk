package git

import (
	"context"
	"strings"
	"time"
)

// Pager builds a snapshot from supplied Git output without rerunning the diff.
// Unsupported patches return an error so the CLI can print the original bytes.
func Pager(ctx context.Context, raw []byte) (*Snapshot, error) {
	files, err := parsePagerFiles(string(raw))
	if err != nil {
		return nil, err
	}
	for i := range files {
		reconstructContext(ctx, &files[i])
		files[i].Patch = normalizePagerPrefixes(files[i].Patch)
		files[i].FullPatch = normalizePagerPrefixes(files[i].FullPatch)
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	root, _ := Command(ctx, "rev-parse", "--show-toplevel").Output()
	snapshot := &Snapshot{
		Repository: strings.TrimSpace(string(root)), Arguments: []string{},
		RepositoryDisplay: repositoryDisplay(strings.TrimSpace(string(root))),
		CreatedAt:         time.Now().UTC().Format(time.RFC3339), Files: sortFiles(files),
	}
	return snapshot, nil
}
