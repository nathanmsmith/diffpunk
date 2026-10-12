package private

import (
	"fmt"
	"strings"
)

func headerNames(s string) (string, string, error) {
	var old, rest string
	if strings.HasPrefix(s, `"`) {
		escaped := false
		for i := 1; i < len(s); i++ {
			if s[i] == '"' && !escaped {
				old, rest = s[:i+1], strings.TrimPrefix(s[i+1:], " ")
				break
			}
			if s[i] == '\\' {
				escaped = !escaped
			} else {
				escaped = false
			}
		}
	} else {
		// Git leaves spaces unquoted; text headers and rename metadata below
		// provide the definitive names for those files.
		split := -1
		delimiter := ""
		for _, prefix := range []string{"b/", "a/", "i/", "w/", "c/", "o/"} {
			for _, candidate := range []string{" " + prefix, ` "` + prefix} {
				if at := strings.Index(s, candidate); at >= 0 && (split < 0 || at < split) {
					split, delimiter = at, candidate
				}
			}
		}
		if split >= 0 {
			old, rest = s[:split], s[split+1:]
			// For ordinary changes, both paths are identical. Check all
			// delimiters because a filename itself may contain " b/".
			for offset := 0; offset < len(s); {
				next := strings.Index(s[offset:], delimiter)
				if next < 0 {
					break
				}
				next += offset
				left, right := s[:next], s[next+1:]
				if patchName(left, true) == patchName(right, true) {
					old, rest = left, right
					break
				}
				offset = next + 1
			}
		}
	}
	if old == "" || rest == "" {
		return "", "", fmt.Errorf("invalid diff header: %s", s)
	}
	return patchName(old, true), patchName(rest, true), nil
}
