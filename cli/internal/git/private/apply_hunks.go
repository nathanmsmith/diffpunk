package private

import (
	"strings"
)

// applyHunks requires exact positions and byte-for-byte matching context. Reverse
// application allows reconstruction when only the new blob is available.
func applyHunks(source string, hunks []patchHunk, reverse bool) (string, bool) {
	lines := splitContent(source)
	result := []string{}
	cursor := 0
	for _, hunk := range hunks {
		start, count := hunk.oldStart, hunk.oldCount
		targetStart, targetCount := hunk.newStart, hunk.newCount
		remove, insert := byte('-'), byte('+')
		if reverse {
			start, count, targetStart, targetCount = targetStart, targetCount, start, count
			remove, insert = insert, remove
		}
		offset := hunkOffset(start, count)
		if offset < cursor || offset > len(lines) || count > len(lines)-offset {
			return "", false
		}
		result = append(result, lines[cursor:offset]...)
		if hunkOffset(targetStart, targetCount) != len(result) {
			return "", false
		}
		cursor = offset
		for _, line := range hunk.lines {
			if line.kind != insert {
				if cursor >= len(lines) || lines[cursor] != line.text {
					return "", false
				}
				cursor++
			}
			if line.kind != remove {
				result = append(result, line.text)
			}
		}
	}
	result = append(result, lines[cursor:]...)
	// A line without a newline can only occur at the end of a text file.
	for _, line := range result[:max(0, len(result)-1)] {
		if !strings.HasSuffix(line, "\n") {
			return "", false
		}
	}
	return strings.Join(result, ""), true
}
