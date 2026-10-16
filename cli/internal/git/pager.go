package git

import (
	"context"
	"strings"
	"time"
)

// Pager builds a snapshot from supplied Git output without rerunning the diff.
// Unsupported patches return an error so the CLI can print the original bytes.
func Pager(ctx context.Context, raw []byte) (*Snapshot, error) {
	files := []File{}
	if len(raw) > 0 {
		var err error
		files, err = parsePagerFiles(string(raw))
		if err != nil {
			return nil, err
		}
	}
	untracked := pagerUntrackedOptions(files)
	if len(raw) == 0 {
		untracked = &UntrackedOptions{DefaultIncluded: true}
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
	if untracked != nil && len(root) > 0 {
		additional, err := loadUntracked(ctx, strings.TrimSpace(string(root)), nil, "", true)
		if err != nil {
			return nil, err
		}
		files = append(files, additional...)
	} else {
		untracked = nil
	}
	snapshot := &Snapshot{
		Repository: strings.TrimSpace(string(root)), Arguments: []string{},
		RepositoryDisplay: repositoryDisplay(strings.TrimSpace(string(root))),
		CreatedAt:         time.Now().UTC().Format(time.RFC3339), Files: sortFiles(files),
		Untracked: untracked,
	}
	return snapshot, nil
}
