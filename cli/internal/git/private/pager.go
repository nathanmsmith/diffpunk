package private

import (
	"regexp"
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
