package git

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func untrackedNames(snapshot *Snapshot) []string {
	names := []string{}
	for _, file := range snapshot.Files {
		if file.Kind == "untracked" {
			names = append(names, file.Name)
		}
	}
	return names
}

func TestUntrackedStandaloneSelection(t *testing.T) {
	testRepo(t)
	writeTest(t, "tracked.txt", "initial\n")
	writeTest(t, ".gitignore", "*.ignored\n")
	gitTest(t, "add", ".")
	gitTest(t, "commit", "-qm", "initial")
	writeTest(t, "new.go", "new\n")
	writeTest(t, "src/new.go", "nested\n")
	writeTest(t, "src/other.txt", "other\n")
	writeTest(t, "hidden.ignored", "ignored\n")
	all := []string{"new.go", "src/new.go", "src/other.txt"}
	for _, tc := range []struct {
		name             string
		args             []string
		hidden, included bool
		want             []string
	}{
		{name: "working tree", want: all},
		{name: "staged", args: []string{"--staged"}, included: true, want: all},
		{name: "cached revision", args: []string{"--cached", "HEAD"}, included: true, want: all},
		{name: "HEAD", args: []string{"HEAD"}, want: all},
		{name: "range", args: []string{"HEAD..HEAD"}, hidden: true, want: []string{}},
		{name: "merge base range", args: []string{"HEAD...HEAD"}, hidden: true, want: []string{}},
		{name: "two revisions", args: []string{"HEAD", "HEAD"}, hidden: true, want: []string{}},
		{name: "path separator", args: []string{"--", "src/"}, want: []string{"src/new.go", "src/other.txt"}},
		{name: "new file path", args: []string{"--", "new.go"}, want: []string{"new.go"}},
		{name: "glob", args: []string{"--", ":(glob)src/*.go"}, want: []string{"src/new.go"}},
		{name: "exclude", args: []string{"--", ":(exclude)src/"}, want: []string{"new.go"}},
		{name: "no match", args: []string{"--", "missing"}, want: []string{}},
		{name: "revision-shaped search", args: []string{"-S", "HEAD"}, want: all},
		{name: "algorithm operand", args: []string{"--diff-algorithm", "minimal"}, want: all},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := takeSnapshot(t, tc.args...)
			if (snapshot.Untracked == nil) != tc.hidden {
				t.Fatalf("untracked options %#v, hidden=%v", snapshot.Untracked, tc.hidden)
			}
			if !tc.hidden && snapshot.Untracked.DefaultIncluded != tc.included {
				t.Fatalf("default included=%v, want %v", snapshot.Untracked.DefaultIncluded, tc.included)
			}
			if got := untrackedNames(snapshot); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("untracked names %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUntrackedRelativePaths(t *testing.T) {
	testRepo(t)
	writeTest(t, "new.txt", "root\n")
	writeTest(t, "src/new.txt", "nested\n")
	if err := os.Chdir("src"); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		args []string
		want []string
	}{
		{nil, []string{"new.txt", "src/new.txt"}},
		{[]string{"--relative"}, []string{"new.txt"}},
		{[]string{"--", "new.txt"}, []string{"src/new.txt"}},
		{[]string{"--relative", "--", "new.txt"}, []string{"new.txt"}},
		{[]string{"--", ":(top)new.txt"}, []string{"new.txt"}},
	} {
		snapshot := takeSnapshot(t, tc.args...)
		if got := untrackedNames(snapshot); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("args %v: got %q, want %q", tc.args, got, tc.want)
		}
	}
	gitTest(t, "config", "diff.relative", "true")
	assertNames(t, takeSnapshot(t), "new.txt")
	assertNames(t, takeSnapshot(t, "--no-relative"), "new.txt", "src/new.txt")
}

func TestUntrackedContentsAndReadOnlyIndex(t *testing.T) {
	root := testRepo(t)
	writeTest(t, "tracked", "tracked\n")
	gitTest(t, "add", ".")
	gitTest(t, "commit", "-qm", "initial")
	indexBefore, err := os.ReadFile(filepath.Join(root, ".git", "index"))
	if err != nil {
		t.Fatal(err)
	}
	contents := map[string]string{
		"empty.txt":       "",
		"binary.dat":      "\x00binary",
		"space name.txt":  "new\nline without newline",
		"tab\tname.txt":   "tab\n",
		"line\nname.txt":  "newline\n",
		"quote\"name.txt": "quote\n",
		"雪.txt":           "snow\n",
		"-dash.txt":       "dash\n",
		"executable":      "#!/bin/sh\n",
	}
	for name, content := range contents {
		writeTest(t, name, content)
	}
	if err := os.Chmod("executable", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing-target", "link"); err != nil {
		t.Fatal(err)
	}
	gitTest(t, "init", "-q", "embedded")
	snapshot := takeSnapshot(t)
	if len(snapshot.Files) != len(contents)+1 {
		t.Fatalf("files: %q", untrackedNames(snapshot))
	}
	for _, file := range snapshot.Files {
		if file.Kind != "untracked" {
			t.Fatalf("kind %q", file.Kind)
		}
		if file.Name == "binary.dat" {
			if !strings.Contains(file.Patch, "Binary files ") || strings.ContainsRune(file.Patch, 0) {
				t.Fatalf("binary patch %q", file.Patch)
			}
			continue
		}
		if file.Name == "link" {
			if !strings.Contains(file.Patch, "new file mode 120000") || !strings.Contains(file.Patch, "+missing-target") {
				t.Fatalf("symlink patch %q", file.Patch)
			}
			continue
		}
		_, hunks, err := parseHunks(file.Patch)
		if err != nil {
			t.Fatal(err)
		}
		got, ok := applyHunks("", hunks, false)
		if !ok || got != contents[file.Name] {
			t.Fatalf("%q content %q, want %q", file.Name, got, contents[file.Name])
		}
	}
	indexAfter, err := os.ReadFile(filepath.Join(root, ".git", "index"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(indexBefore, indexAfter) {
		t.Fatal("untracked loading changed the index")
	}
	writeTest(t, "space name.txt", "changed\n")
	refreshed := takeSnapshot(t)
	for _, file := range refreshed.Files {
		if file.Name == "space name.txt" && !strings.Contains(file.Patch, "+changed") {
			t.Fatal("refresh missed untracked edit")
		}
	}
	for _, file := range snapshot.Files {
		if file.Name == "space name.txt" && strings.Contains(file.Patch, "+changed") {
			t.Fatal("snapshot changed after editing file")
		}
	}
}

func TestPagerUntrackedModesAndEmptyInput(t *testing.T) {
	testRepo(t)
	writeTest(t, "tracked.txt", "initial\n")
	gitTest(t, "add", ".")
	gitTest(t, "commit", "-qm", "initial")
	writeTest(t, "tracked.txt", "committed\n")
	gitTest(t, "add", "tracked.txt")
	gitTest(t, "commit", "-qm", "second")
	writeTest(t, "tracked.txt", "staged\n")
	gitTest(t, "add", "tracked.txt")
	writeTest(t, "tracked.txt", "working\n")
	writeTest(t, "new.txt", "untracked\n")
	writeTest(t, "src/new.txt", "nested\n")
	for _, tc := range []struct {
		name             string
		args             []string
		hidden, included bool
	}{
		{name: "working"},
		{name: "staged", args: []string{"--staged"}, included: true},
		{name: "cached", args: []string{"--cached"}, included: true},
		{name: "HEAD", args: []string{"HEAD"}},
		{name: "reverse staged", args: []string{"--staged", "-R"}, included: true},
		{name: "unknown prefixes", args: []string{"--default-prefix"}, hidden: true},
		{name: "historical", args: []string{"HEAD~1", "HEAD"}, hidden: true},
		{name: "empty historical", args: []string{"HEAD", "HEAD"}, included: true},
		{name: "path filtered", args: []string{"--", "tracked.txt"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			patch := gitTest(t, append([]string{"-c", "diff.mnemonicPrefix=true", "diff", "--no-color"}, tc.args...)...)
			snapshot, err := Pager(context.Background(), []byte(patch))
			if err != nil {
				t.Fatal(err)
			}
			if (snapshot.Untracked == nil) != tc.hidden {
				t.Fatalf("untracked %#v", snapshot.Untracked)
			}
			if tc.hidden {
				if len(untrackedNames(snapshot)) != 0 {
					t.Fatal("unknown comparison included untracked files")
				}
			} else if snapshot.Untracked.DefaultIncluded != tc.included || !reflect.DeepEqual(untrackedNames(snapshot), []string{"new.txt", "src/new.txt"}) {
				t.Fatalf("untracked %#v, files %q", snapshot.Untracked, untrackedNames(snapshot))
			}
		})
	}
	if err := os.Chdir("src"); err != nil {
		t.Fatal(err)
	}
	snapshot, err := Pager(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	assertNames(t, snapshot, "new.txt", "src/new.txt")
	if err := os.Remove("new.txt"); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove("../new.txt"); err != nil {
		t.Fatal(err)
	}
	snapshot, err = Pager(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Files) != 0 {
		t.Fatalf("empty pager files %v", snapshot.Files)
	}
	patch := gitTest(t, "-c", "diff.mnemonicPrefix=true", "diff", "--staged", "--no-color")
	snapshot, err = Pager(context.Background(), []byte(patch))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Untracked == nil || !snapshot.Untracked.DefaultIncluded {
		t.Fatal("staged toggle unavailable without untracked files")
	}
}

func TestUntrackedRefreshLifecycle(t *testing.T) {
	testRepo(t)
	initial := takeSnapshot(t)
	if initial.Untracked == nil || initial.Untracked.DefaultIncluded || len(initial.Files) != 0 {
		t.Fatalf("initial snapshot %#v", initial)
	}
	writeTest(t, "new.txt", "first\n")
	created := takeSnapshot(t)
	assertNames(t, created, "new.txt")
	writeTest(t, "new.txt", "second\n")
	if !strings.Contains(takeSnapshot(t).Files[0].Patch, "+second\n") {
		t.Fatal("refresh did not load changed untracked contents")
	}
	if !strings.Contains(created.Files[0].Patch, "+first\n") || len(initial.Files) != 0 {
		t.Fatal("earlier snapshots changed")
	}
	gitTest(t, "add", "new.txt")
	staged := takeSnapshot(t, "--staged")
	if staged.Files[0].Kind != "added" || len(untrackedNames(staged)) != 0 {
		t.Fatal("staged file is still marked untracked")
	}
	if len(takeSnapshot(t).Files) != 0 {
		t.Fatal("unchanged staged file appeared in working-tree comparison")
	}
	writeTest(t, "deleted.txt", "temporary\n")
	assertNames(t, takeSnapshot(t), "deleted.txt")
	if err := os.Remove("deleted.txt"); err != nil {
		t.Fatal(err)
	}
	if len(takeSnapshot(t).Files) != 0 {
		t.Fatal("removed untracked file survived refresh")
	}
}

func TestStagedDeletionAndUntrackedReplacement(t *testing.T) {
	testRepo(t)
	writeTest(t, "same.txt", "original\n")
	gitTest(t, "add", ".")
	gitTest(t, "commit", "-qm", "initial")
	gitTest(t, "rm", "--cached", "same.txt")
	writeTest(t, "same.txt", "replacement\n")
	snapshot := takeSnapshot(t, "--staged")
	if len(snapshot.Files) != 2 || snapshot.Files[0].Kind != "deleted" || snapshot.Files[1].Kind != "untracked" {
		t.Fatalf("staged deletion and replacement: %#v", snapshot.Files)
	}
	if !strings.Contains(snapshot.Files[0].Patch, "-original\n") || !strings.Contains(snapshot.Files[1].Patch, "+replacement\n") {
		t.Fatal("staged deletion and untracked contents were mixed")
	}
}
