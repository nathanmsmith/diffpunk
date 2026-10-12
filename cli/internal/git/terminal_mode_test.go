package git

import (
	"testing"
)

func TestTerminalModesAndPathSeparator(t *testing.T) {
	for _, arg := range []string{"--stat", "--stat=80", "--name-only", "--check", "--word-diff=color", "--raw", "--quiet", "--exit-code", "--binary"} {
		if !TerminalMode([]string{arg}) {
			t.Errorf("%s should pass through", arg)
		}
		if TerminalMode([]string{"--", arg}) {
			t.Errorf("%s is a pathspec", arg)
		}
	}
	for _, arg := range []string{"--staged", "-U20", "-w", "HEAD~1..HEAD", "--find-renames"} {
		if TerminalMode([]string{arg}) {
			t.Errorf("%s should render", arg)
		}
	}
}

func TestTerminalStatOptions(t *testing.T) {
	for _, arg := range []string{"--stat-width=80", "--stat-name-width=40", "--stat-count=5"} {
		if !TerminalMode([]string{arg}) {
			t.Errorf("%s should pass through", arg)
		}
		if TerminalMode([]string{"--", arg}) {
			t.Errorf("%s is a pathspec", arg)
		}
	}
}
