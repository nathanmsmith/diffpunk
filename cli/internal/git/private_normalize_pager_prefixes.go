package git

import (
	"regexp"
	"strings"
)

var pagerFileHeader = regexp.MustCompile(`^diff --git ("[abiwco]/(?:[^"\\]|\\.)*"|[abiwco]/.+?) ("[abiwco]/(?:[^"\\]|\\.)*"|[abiwco]/.+?)$`)

// The browser's diff parser requires a/ and b/ transport prefixes. Git can
// supply mnemonic prefixes in pager mode; normalize only file headers, leaving
// the supplied hunk bodies and unprefixed rename/copy metadata intact.
func normalizePagerPrefixes(patch string) string {
	lines := strings.SplitAfter(patch, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "@@") {
			break
		}
		text := strings.TrimSuffix(line, "\n")
		newline := line[len(text):]
		switch {
		case strings.HasPrefix(text, "diff --git "):
			if names := pagerFileHeader.FindStringSubmatch(text); names != nil {
				lines[i] = "diff --git " + pagerPathPrefix(names[1], "a") + " " + pagerPathPrefix(names[2], "b") + newline
			}
		case strings.HasPrefix(text, "--- "), strings.HasPrefix(text, "+++ "):
			prefix := "a"
			if strings.HasPrefix(text, "+++ ") {
				prefix = "b"
			}
			lines[i] = text[:4] + pagerPathPrefix(text[4:], prefix) + newline
		}
	}
	return strings.Join(lines, "")
}
