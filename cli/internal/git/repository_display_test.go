package git

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestRepositoryDisplay(t *testing.T) {
	home := filepath.Join(t.TempDir(), "user")
	variable := "HOME"
	if runtime.GOOS == "windows" {
		variable = "USERPROFILE"
	}
	t.Setenv(variable, home)
	for _, tt := range []struct {
		path, want string
	}{
		{home, "~"},
		{filepath.Join(home, "Developer", "diffpunk"), "~" + string(filepath.Separator) + filepath.Join("Developer", "diffpunk")},
		{home + "-other", home + "-other"},
		{filepath.Dir(home), filepath.Dir(home)},
		{"", ""},
	} {
		if got := repositoryDisplay(tt.path); got != tt.want {
			t.Errorf("repositoryDisplay(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
	t.Setenv(variable, "")
	if got := repositoryDisplay(home); got != home {
		t.Errorf("without home: got %q, want %q", got, home)
	}
}
