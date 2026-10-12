package private

import (
	"strings"
)

func pagerPathPrefix(name, prefix string) string {
	offset := 0
	if strings.HasPrefix(name, `"`) {
		offset = 1
	}
	if len(name) > offset+1 && name[offset+1] == '/' && strings.ContainsRune("abiwco", rune(name[offset])) {
		return name[:offset] + prefix + name[offset+1:]
	}
	return name
}
