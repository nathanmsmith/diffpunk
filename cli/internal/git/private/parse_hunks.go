package private

import (
	"fmt"
	"strconv"
	"strings"
)

// parseHunks rejects word/combined diffs and malformed counts before the browser
// is opened. Binary and metadata-only changes have no text hunks.
func parseHunks(patch string) (string, []patchHunk, error) {
	lines := strings.SplitAfter(patch, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	prefix := ""
	hunks := []patchHunk{}
	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			match := hunkHeader.FindStringSubmatch(line)
			if match == nil {
				return "", nil, fmt.Errorf("unsupported hunk header")
			}
			values := [4]int{}
			for i := range values {
				if match[i+1] == "" {
					values[i] = 1
					continue
				}
				n, err := strconv.Atoi(match[i+1])
				if err != nil {
					return "", nil, err
				}
				values[i] = n
			}
			hunks = append(hunks, patchHunk{oldStart: values[0], oldCount: values[1], newStart: values[2], newCount: values[3]})
			continue
		}
		if len(hunks) == 0 {
			prefix += line
			continue
		}
		hunk := &hunks[len(hunks)-1]
		if line == "\\ No newline at end of file\n" || line == "\\ No newline at end of file" {
			if len(hunk.lines) == 0 || !strings.HasSuffix(hunk.lines[len(hunk.lines)-1].text, "\n") {
				return "", nil, fmt.Errorf("invalid no-newline marker")
			}
			hunk.lines[len(hunk.lines)-1].text = strings.TrimSuffix(hunk.lines[len(hunk.lines)-1].text, "\n")
			continue
		}
		if len(line) == 0 || (line[0] != ' ' && line[0] != '-' && line[0] != '+') {
			return "", nil, fmt.Errorf("unsupported patch body")
		}
		hunk.lines = append(hunk.lines, patchLine{kind: line[0], text: line[1:]})
	}
	for _, hunk := range hunks {
		oldCount, newCount := 0, 0
		for _, line := range hunk.lines {
			if line.kind != '+' {
				oldCount++
			}
			if line.kind != '-' {
				newCount++
			}
		}
		if oldCount != hunk.oldCount || newCount != hunk.newCount ||
			(hunk.oldCount > 0 && hunk.oldStart == 0) || (hunk.newCount > 0 && hunk.newStart == 0) {
			return "", nil, fmt.Errorf("invalid hunk counts")
		}
	}
	return prefix, hunks, nil
}
