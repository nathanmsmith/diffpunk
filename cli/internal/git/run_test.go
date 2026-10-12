package git

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func testRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	gitTest(t, "init", "-q")
	gitTest(t, "config", "user.name", "Diff Punk Test")
	gitTest(t, "config", "user.email", "test@example.invalid")
	gitTest(t, "config", "commit.gpgsign", "false")
	gitTest(t, "config", "core.autocrlf", "false")
	gitTest(t, "config", "diff.mnemonicPrefix", "false")
	gitTest(t, "config", "diff.noprefix", "false")
	return root
}

func gitTest(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func writeTest(t *testing.T, name, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}

func takeSnapshot(t *testing.T, args ...string) *Snapshot {
	t.Helper()
	snapshot, raw, err := Run(context.Background(), args, os.Stderr)
	if err != nil || raw != "" || snapshot == nil {
		t.Fatalf("run %v: %v, raw %q, snapshot %v", args, err, raw, snapshot)
	}
	return snapshot
}

func assertNames(t *testing.T, snapshot *Snapshot, want ...string) {
	t.Helper()
	got := []string{}
	for _, f := range snapshot.Files {
		got = append(got, f.Name)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("files %v, want %v", got, want)
	}
}

func TestGitSelectionAndFrozenContext(t *testing.T) {
	testRepo(t)
	lines := []string{}
	for i := 0; i < 30; i++ {
		lines = append(lines, "unchanged context")
	}
	lines[15] = "original"
	original := strings.Join(lines, "\n") + "\n"
	writeTest(t, "foo.go", original)
	writeTest(t, "foo_test.go", "original test\n")
	writeTest(t, "nested/other.go", "original other\n")
	gitTest(t, "add", ".")
	gitTest(t, "commit", "-qm", "initial")
	lines[15] = "staged"
	writeTest(t, "foo.go", strings.Join(lines, "\n")+"\n")
	gitTest(t, "add", "foo.go")
	lines[15] = "working"
	writeTest(t, "foo.go", strings.Join(lines, "\n")+"\n")
	writeTest(t, "foo_test.go", "changed test\n")
	writeTest(t, "nested/other.go", "changed other\n")
	writeTest(t, "untracked.go", "not included\n")
	unstaged := takeSnapshot(t)
	assertNames(t, unstaged, "foo.go", "foo_test.go", "nested/other.go")
	if !strings.Contains(unstaged.Files[0].Patch, "-staged\n+working") {
		t.Fatal("default did not compare index to working tree")
	}
	if strings.Count(unstaged.Files[0].FullPatch, "unchanged context") != 29 {
		t.Fatal("full context is incomplete")
	}
	staged := takeSnapshot(t, "--staged")
	assertNames(t, staged, "foo.go")
	if !strings.Contains(staged.Files[0].Patch, "-original\n+staged") {
		t.Fatal("staged did not compare HEAD to index")
	}
	head := takeSnapshot(t, "HEAD", "--", "foo.go")
	assertNames(t, head, "foo.go")
	if !strings.Contains(head.Files[0].Patch, "-original\n+working") {
		t.Fatal("HEAD diff incorrect")
	}
	writeTest(t, "foo.go", "edited after snapshot\n")
	if strings.Contains(unstaged.Files[0].FullPatch, "edited after snapshot") {
		t.Fatal("snapshot changed")
	}
}

func TestRevisionRangeAndRelativePaths(t *testing.T) {
	testRepo(t)
	writeTest(t, "src/foo.ts", "first\n")
	gitTest(t, "add", ".")
	gitTest(t, "commit", "-qm", "first")
	writeTest(t, "src/foo.ts", "second\n")
	gitTest(t, "add", ".")
	gitTest(t, "commit", "-qm", "second")
	snapshot := takeSnapshot(t, "HEAD~1..HEAD")
	assertNames(t, snapshot, "src/foo.ts")
	if !strings.Contains(snapshot.Files[0].FullPatch, "-first\n+second") {
		t.Fatal("historical content incorrect")
	}
	writeTest(t, "src/foo.ts", "third\n")
	if err := os.Chdir("src"); err != nil {
		t.Fatal(err)
	}
	snapshot = takeSnapshot(t, "--relative", "--", "foo.ts")
	assertNames(t, snapshot, "foo.ts")
}

func TestFileKindsAndQuotedPaths(t *testing.T) {
	testRepo(t)
	for name, content := range map[string]string{"old.txt": "renamed\n", "deleted.txt": "deleted\n", "binary.dat": "\x00original", "folder b/nested.bin": "\x00original", "space name.txt": "original\n", "雪.txt": "original\n", "tab\tname.txt": "original\n", "mode.txt": "unchanged\n"} {
		writeTest(t, name, content)
	}
	gitTest(t, "add", ".")
	gitTest(t, "commit", "-qm", "initial")
	gitTest(t, "mv", "old.txt", "new.txt")
	if err := os.Remove("deleted.txt"); err != nil {
		t.Fatal(err)
	}
	writeTest(t, "binary.dat", "\x00changed")
	writeTest(t, "folder b/nested.bin", "\x00changed")
	writeTest(t, "space name.txt", "changed\n")
	writeTest(t, "雪.txt", "changed\n")
	writeTest(t, "tab\tname.txt", "changed\n")
	writeTest(t, "added.txt", "added\n")
	if err := os.Chmod("mode.txt", 0755); err != nil {
		t.Fatal(err)
	}
	gitTest(t, "add", ".")
	snapshot := takeSnapshot(t, "--cached", "-M")
	kinds := map[string]string{}
	for _, f := range snapshot.Files {
		kinds[f.Name] = f.Kind
		if f.Name == "new.txt" && (f.FullContent == nil || *f.FullContent != "renamed\n") {
			t.Fatal("rename full content unavailable")
		}
		if f.Name == "mode.txt" && (f.FullContent == nil || *f.FullContent != "unchanged\n") {
			t.Fatal("mode change full content unavailable")
		}
	}
	want := map[string]string{"new.txt": "renamed", "deleted.txt": "deleted", "binary.dat": "modified", "folder b/nested.bin": "modified", "space name.txt": "modified", "雪.txt": "modified", "tab\tname.txt": "modified", "added.txt": "added", "mode.txt": "modified"}
	if !reflect.DeepEqual(kinds, want) {
		t.Fatalf("got %v; want %v", kinds, want)
	}
}

func TestGitOutputPreserved(t *testing.T) {
	testRepo(t)
	writeTest(t, "sample.txt", "one\ntwo\nthree\nfour\nfive\n")
	gitTest(t, "add", ".")
	gitTest(t, "commit", "-qm", "initial")
	writeTest(t, "sample.txt", "one\ntwo\nTHREE\nfour\nfive\n")
	for _, args := range [][]string{{}, {"-U0"}, {"-w"}, {"--patience"}, {"--", "sample.txt"}} {
		snapshot := takeSnapshot(t, args...)
		native := gitTest(t, append([]string{"diff", "--no-color"}, args...)...)
		if snapshot.Files[0].Patch != native {
			t.Fatalf("args %v patch differs from git:\n%q\n%q", args, snapshot.Files[0].Patch, native)
		}
	}
}

func TestEmptyDiff(t *testing.T) {
	testRepo(t)
	snapshot, raw, err := Run(context.Background(), nil, os.Stderr)
	if err != nil || snapshot != nil || raw != "" {
		t.Fatalf("empty diff: %v %q %v", snapshot, raw, err)
	}
}

func TestRunDiagnostics(t *testing.T) {
	testRepo(t)
	var diagnostics bytes.Buffer
	snapshot, raw, err := Run(context.Background(), []string{"missing-revision"}, &diagnostics)
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || snapshot != nil || raw != "" {
		t.Fatalf("failed comparison: %v %q %v", snapshot, raw, err)
	}
	if !strings.Contains(diagnostics.String(), "missing-revision") {
		t.Fatalf("missing Git diagnostics: %q", diagnostics.String())
	}
}
