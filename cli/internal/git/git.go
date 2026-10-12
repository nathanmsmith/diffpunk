package git

import (
	"regexp"
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
	Repository string   `json:"repository"`
	Arguments  []string `json:"arguments"`
	CreatedAt  string   `json:"createdAt"`
	Files      []File   `json:"files"`
}

var fileStart = regexp.MustCompile(`(?m)^diff --git `)
