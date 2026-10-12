package git

import (
	"reflect"
	"testing"
)

func TestParseFilesRejectsUnsupportedHeaders(t *testing.T) {
	for _, patch := range []string{"", "summary\ndiff --git a/f b/f\n", "diff --git invalid\n", "diff --git a/f b/f\n\ndiff --cc f\n", "diff --git a/f b/f\n\ndiff --combined f\n"} {
		if _, err := parseFiles(patch); err == nil {
			t.Errorf("accepted %q", patch)
		}
	}
}

func TestParseCopiedFile(t *testing.T) {
	patch := "diff --git a/old b/new\nsimilarity index 100%\ncopy from old\ncopy to new\n"
	got, err := parseFiles(patch)
	want := []File{{Name: "new", PreviousName: "old", Kind: "copied", Patch: patch}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v %v; want %#v", got, err, want)
	}
}

func TestHeaderNames(t *testing.T) {
	for _, tt := range []struct{ header, old, new string }{
		{`"a/quote\"name" "b/quote\"name"`, `quote"name`, `quote"name`},
		{`a/space name "b/tab\tname"`, "space name", "tab\tname"},
		{"a/folder b/name b/folder b/name", "folder b/name", "folder b/name"},
		{"i/old w/new", "old", "new"},
	} {
		old, newName, err := headerNames(tt.header)
		if err != nil || old != tt.old || newName != tt.new {
			t.Errorf("%q: got %q %q %v", tt.header, old, newName, err)
		}
	}
	for _, header := range []string{`"unterminated`, "a/name", `"a/name" `} {
		if _, _, err := headerNames(header); err == nil {
			t.Errorf("accepted %q", header)
		}
	}
}

func TestNormalizePagerPrefixes(t *testing.T) {
	patch := "diff --git i/name w/name\n--- i/name\n+++ w/name\n@@ -1 +1 @@\n--- i/body\n+++ w/body\n"
	want := "diff --git a/name b/name\n--- a/name\n+++ b/name\n@@ -1 +1 @@\n--- i/body\n+++ w/body\n"
	if got := normalizePagerPrefixes(patch); got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
	for _, tt := range []struct{ patch, want string }{
		{`diff --git "c/tab\tname" "o/tab\tname"`, `diff --git "a/tab\tname" "b/tab\tname"`},
		{"--- /dev/null\n+++ w/name", "--- /dev/null\n+++ b/name"},
		{"diff --git no-prefix invalid\nrename from i/name\nrename to w/name\n", "diff --git no-prefix invalid\nrename from i/name\nrename to w/name\n"},
	} {
		if got := normalizePagerPrefixes(tt.patch); got != tt.want {
			t.Errorf("got %q; want %q", got, tt.want)
		}
	}
	for _, name := range []string{"", "x", "name", "/dev/null", "z/name"} {
		if got := pagerPathPrefix(name, "a"); got != name {
			t.Errorf("changed %q to %q", name, got)
		}
	}
}
