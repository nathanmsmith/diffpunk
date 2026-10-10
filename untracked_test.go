package main

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestUntrackedSnapshot(t *testing.T) {
	testRepo(t)
	writeTest(t, ".gitignore", "ignored/\n")
	gitTest(t, "add", ".gitignore")
	gitTest(t, "commit", "-qm", "first")
	gitTest(t, "commit", "--allow-empty", "-qm", "second")
	writeTest(t, "new.go", "package example\n")
	writeTest(t, "new_test.go", "package example\n")
	writeTest(t, "src/雪\tname.txt", "no newline")
	writeTest(t, "empty.txt", "")
	writeTest(t, "binary.dat", "\x00binary")
	writeTest(t, "ignored/hidden.txt", "ignored\n")
	if err := os.Symlink("new.go", "link"); err != nil {
		t.Fatal(err)
	}
	before := gitTest(t, "status", "--porcelain=v1", "-z")
	snapshot := takeSnapshot(t)
	if !snapshot.UntrackedAvailable || snapshot.ShowUntracked {
		t.Fatal("incorrect defaults")
	}
	if len(snapshot.Files) != 6 {
		t.Fatalf("got %d files: %v", len(snapshot.Files), snapshot.Files)
	}
	for _, file := range snapshot.Files {
		if file.Kind != "untracked" {
			t.Fatalf("incorrect kind: %v", file)
		}
		switch file.Name {
		case "new.go":
			if !strings.Contains(file.Patch, "+package example") {
				t.Fatal("missing content")
			}
		case "binary.dat":
			if !strings.Contains(file.Patch, "Binary files") {
				t.Fatal("missing binary summary")
			}
		case "link":
			if !strings.Contains(file.Patch, "120000") || !strings.Contains(file.Patch, "+new.go") {
				t.Fatal("symlink not captured")
			}
		}
	}
	shown := takeSnapshot(t, "--untracked")
	if !shown.ShowUntracked || len(shown.Arguments) != 0 {
		t.Fatal("CLI flag not consumed")
	}
	assertNames(t, takeSnapshot(t, "HEAD", "--", "src/"), "src/雪\tname.txt")
	assertNames(t, takeSnapshot(t, "-S", "package", "--", "new.go"), "new.go")
	if err := os.Chdir("src"); err != nil {
		t.Fatal(err)
	}
	assertNames(t, takeSnapshot(t, "--relative"), "雪\tname.txt")
	if err := os.Chdir(".."); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"--staged"}, {"--cached", "--untracked"}, {"HEAD~1", "HEAD"}, {"HEAD~1..HEAD"}, {"HEAD~1...HEAD"}} {
		empty, raw, err := capture(context.Background(), args)
		if err != nil || raw != "" || empty != nil {
			t.Fatalf("%v included untracked files: %v %v", args, empty, err)
		}
	}
	writeTest(t, "new.go", "edited later\n")
	for _, file := range snapshot.Files {
		if strings.Contains(file.Patch, "edited later") {
			t.Fatal("snapshot changed")
		}
	}
	after := gitTest(t, "status", "--porcelain=v1", "-z")
	if before != after {
		t.Fatal("capture changed index or working tree status")
	}
}

func TestUntrackedExcludedFromStagedAndHistoricalViews(t *testing.T) {
	testRepo(t)
	writeTest(t, "tracked.txt", "first\n")
	gitTest(t, "add", ".")
	gitTest(t, "commit", "-qm", "first")
	writeTest(t, "tracked.txt", "second\n")
	gitTest(t, "add", ".")
	gitTest(t, "commit", "-qm", "second")
	writeTest(t, "tracked.txt", "staged\n")
	gitTest(t, "add", "tracked.txt")
	writeTest(t, "new.txt", "untracked\n")
	for _, args := range [][]string{{"--untracked", "--staged"}, {"--untracked", "HEAD~1", "HEAD"}} {
		snapshot := takeSnapshot(t, args...)
		assertNames(t, snapshot, "tracked.txt")
		if snapshot.UntrackedAvailable || snapshot.ShowUntracked {
			t.Fatalf("toggle available for %v", args)
		}
	}
	args, show := viewerArgs([]string{"--untracked", "--", "--untracked"})
	if !show || !reflect.DeepEqual(args, []string{"--", "--untracked"}) {
		t.Fatal("path separator not respected")
	}
}

func TestUntrackedOnlyBeforeFirstCommit(t *testing.T) {
	testRepo(t)
	writeTest(t, "--untracked", "new file\n")
	snapshot := takeSnapshot(t, "--untracked", "--", "--untracked")
	assertNames(t, snapshot, "--untracked")
	if !snapshot.UntrackedAvailable || !snapshot.ShowUntracked {
		t.Fatal("untracked-only view unavailable")
	}
}
