package git

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var blobHeader = regexp.MustCompile(`(?m)^index ([0-9a-f]{4,64})\.\.([0-9a-f]{4,64})(?: [0-7]+)?$`)

func reconstructContext(ctx context.Context, file *File) {
	if strings.Contains(file.Patch, "\nBinary files ") || strings.Contains(file.Patch, "\nGIT binary patch") {
		return
	}
	prefix, hunks, err := parseHunks(file.Patch)
	if err != nil {
		return
	}
	ids := blobHeader.FindStringSubmatch(prefix)
	if ids == nil {
		return
	}
	old, oldOK := readBlob(ctx, ids[1])
	newContent, newOK := readBlob(ctx, ids[2])
	if file.Kind == "added" && strings.Trim(ids[1], "0") == "" {
		old, oldOK = "", true
	}
	if file.Kind == "deleted" && strings.Trim(ids[2], "0") == "" {
		newContent, newOK = "", true
	}
	if len(hunks) == 0 {
		if oldOK && newOK && old == newContent {
			file.FullContent = &old
		}
		return
	}
	if oldOK {
		built, ok := applyHunks(old, hunks, false)
		if !ok || (newOK && built != newContent) || !matchesBlob(ctx, built, ids[2]) {
			return
		}
		newContent, newOK = built, true
	} else if newOK {
		built, ok := applyHunks(newContent, hunks, true)
		if !ok || !matchesBlob(ctx, built, ids[1]) {
			return
		}
		old, oldOK = built, true
	}
	if !oldOK || !newOK || strings.ContainsRune(old, 0) || strings.ContainsRune(newContent, 0) {
		return
	}
	oldLines, newLines := splitContent(old), splitContent(newContent)
	covered := 0
	for _, hunk := range hunks {
		covered += hunk.oldCount
	}
	if covered == len(oldLines) {
		return // The supplied patch already contains the entire file.
	}
	// Preserve all supplied changed lines. Only add verified unchanged context;
	// recomputing a diff could discard the user's chosen algorithm or filters.
	var full strings.Builder
	full.WriteString(prefix)
	oldStart, newStart := 1, 1
	if len(oldLines) == 0 {
		oldStart = 0
	}
	if len(newLines) == 0 {
		newStart = 0
	}
	fmt.Fprintf(&full, "@@ -%d,%d +%d,%d @@\n", oldStart, len(oldLines), newStart, len(newLines))
	cursor := 0
	for _, hunk := range hunks {
		offset := hunkOffset(hunk.oldStart, hunk.oldCount)
		for _, line := range oldLines[cursor:offset] {
			writePatchLine(&full, patchLine{kind: ' ', text: line})
		}
		for _, line := range hunk.lines {
			writePatchLine(&full, line)
		}
		cursor = offset + hunk.oldCount
	}
	for _, line := range oldLines[cursor:] {
		writePatchLine(&full, patchLine{kind: ' ', text: line})
	}
	file.FullPatch = full.String()
}
