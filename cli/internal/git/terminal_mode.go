package git

import (
	"strings"
)

// TerminalMode reports whether the arguments require native Git terminal output.
func TerminalMode(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			break
		}
		name := strings.SplitN(arg, "=", 2)[0]
		switch name {
		case "--stat", "--numstat", "--shortstat", "--dirstat", "--dirstat-by-file", "--cumulative", "--summary", "--compact-summary", "--name-only", "--name-status", "--check", "--raw", "--patch-with-raw", "--patch-with-stat", "--word-diff", "--word-diff-regex", "--color-words", "--no-patch", "--quiet", "--exit-code", "--help", "-h", "-s", "--output", "--binary", "--cc", "-c", "--combined-all-paths", "--ext-diff", "--no-index", "--submodule", "--line-prefix", "--src-prefix", "--dst-prefix", "--no-prefix", "--default-prefix", "--output-indicator-new", "--output-indicator-old", "--output-indicator-context", "--irreversible-delete", "-D":
			return true
		}
		if strings.HasPrefix(name, "--stat-") || strings.HasPrefix(name, "--dirstat=") {
			return true
		}
	}
	return false
}
