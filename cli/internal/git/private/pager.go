package private

import (
	"context"
	"regexp"
	"strings"
	"time"
)

var (
	terminalColor   = regexp.MustCompile(`\x1b\[[0-9;:]*m`)
	pagerFileHeader = regexp.MustCompile(`^diff --git ("[abiwco]/(?:[^"\\]|\\.)*"|[abiwco]/.+?) ("[abiwco]/(?:[^"\\]|\\.)*"|[abiwco]/.+?)$`)
	blobHeader      = regexp.MustCompile(`(?m)^index ([0-9a-f]{4,64})\.\.([0-9a-f]{4,64})(?: [0-7]+)?$`)
	hunkHeader      = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@[^\n]*\n?$`)
)

type patchLine struct {
	kind byte
	text string // Includes the newline, unless Git supplied a no-newline marker.
}

type patchHunk struct {
	oldStart, oldCount int
	newStart, newCount int
	lines              []patchLine
}

// Pager builds a snapshot from supplied Git output without rerunning the diff.
// Unsupported patches return an error so the CLI can print the original bytes.
func Pager(ctx context.Context, raw []byte) (*Snapshot, error) {
	files, err := parseFiles(terminalColor.ReplaceAllString(string(raw), ""))
	if err == nil {
		for _, file := range files {
			if _, _, parseErr := parseHunks(file.Patch); parseErr != nil {
				err = parseErr
				break
			}
		}
	}
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
		CreatedAt: time.Now().UTC().Format(time.RFC3339), Files: sortFiles(files),
	}
	return snapshot, nil
}
