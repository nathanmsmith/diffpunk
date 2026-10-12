package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

// Run runs git diff and builds a snapshot with full-file context.
// Git diagnostics are written to diagnostics.
// Unsupported patch formats are returned as raw output; an empty diff has an empty snapshot.
func Run(ctx context.Context, args []string, diagnostics io.Writer) (*Snapshot, string, error) {
	cmd := Command(ctx, patchArgs(args, false)...)
	cmd.Stderr = diagnostics
	patch, err := cmd.Output()
	if err != nil {
		return nil, "", err
	}
	files := []File{}
	if len(patch) > 0 {
		files, err = parseFiles(string(patch))
		if err != nil {
			return nil, string(patch), nil
		}
	}
	// A second patch uses the same selection and diff algorithm, but includes all
	// context. This also works for staged and historical files and textconv.
	fullCmd := Command(ctx, patchArgs(args, true)...)
	var fullErr bytes.Buffer
	fullCmd.Stderr = &fullErr
	full, err := fullCmd.Output()
	if err == nil {
		expanded, parseErr := parseFiles(string(full))
		if parseErr == nil {
			byName := make(map[string]string, len(expanded))
			for _, f := range expanded {
				byName[f.Name] = f.Patch
			}
			for i := range files {
				files[i].FullPatch = byName[files[i].Name]
			}
		}
	} else {
		_, _ = fmt.Fprintln(diagnostics, "diff punk: full-file context unavailable:", strings.TrimSpace(fullErr.String()))
	}
	loadMetadataContents(ctx, args, files)
	rootCmd := Command(ctx, "rev-parse", "--show-toplevel")
	root, err := rootCmd.Output()
	if err != nil {
		return nil, "", err
	}
	return &Snapshot{
		Repository: strings.TrimSpace(string(root)), Arguments: append([]string{}, args...),
		RepositoryDisplay: repositoryDisplay(strings.TrimSpace(string(root))),
		CreatedAt:         time.Now().UTC().Format(time.RFC3339),
		Files:             sortFiles(files),
	}, "", nil
}
