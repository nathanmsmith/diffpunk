package git

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// File holds a file change and any available full-file context.
type File struct {
	Name         string  `json:"name"`
	PreviousName string  `json:"previousName,omitempty"`
	Kind         string  `json:"kind"`
	Patch        string  `json:"patch"`
	FullPatch    string  `json:"fullPatch,omitempty"`
	FullContent  *string `json:"fullContent,omitempty"`
}

// Snapshot is the frozen diff sent to the browser viewer.
type Snapshot struct {
	Repository        string            `json:"repository"`
	RepositoryDisplay string            `json:"repositoryDisplay,omitempty"`
	Arguments         []string          `json:"arguments"`
	CreatedAt         string            `json:"createdAt"`
	Files             []File            `json:"files"`
	Untracked         *UntrackedOptions `json:"untracked,omitempty"`
}

// UntrackedOptions is present when the viewer can include untracked files.
type UntrackedOptions struct {
	DefaultIncluded bool `json:"defaultIncluded"`
}

func repositoryDisplay(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	home = filepath.Clean(home)
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+string(filepath.Separator)) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

var fileStart = regexp.MustCompile(`(?m)^diff --git `)
