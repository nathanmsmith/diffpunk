package git

import (
	"strconv"
	"strings"
)

func patchName(s string, prefixed bool) string {
	s = strings.TrimSuffix(s, "\r")
	if strings.HasPrefix(s, `"`) {
		if decoded, err := strconv.Unquote(s); err == nil {
			s = decoded
		}
	} else {
		s = strings.TrimSuffix(s, "\t")
	}
	// Git's mnemonic prefixes use i (index), w (working tree), c (commit),
	// and o (object). Direct Run still forces a/ and b/.
	if prefixed && len(s) >= 2 && s[1] == '/' && strings.ContainsRune("abiwco", rune(s[0])) {
		s = s[2:]
	}
	return s
}
