package git

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func storeTestBlob(t *testing.T, content string) string {
	t.Helper()
	cmd := Command(context.Background(), "hash-object", "-w", "--stdin")
	cmd.Stdin = strings.NewReader(content)
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func blobPatch(oldID, newID, body string) string {
	return fmt.Sprintf("diff --git i/file.txt w/file.txt\nindex %s..%s 100644\n--- i/file.txt\n+++ w/file.txt\n%s", oldID, newID, body)
}

func TestPagerReconstructsFrozenContext(t *testing.T) {
	root := testRepo(t)
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	old := "before\nold\nafter\nlast"
	newContent := "before\nnew\nafter\nlast"
	oldID := storeTestBlob(t, old)
	newID := storeTestBlob(t, newContent)
	raw := blobPatch(oldID, newID, "@@ -2 +2 @@\n-old\n+new\n")
	colored := "\x1b[1;31m" + raw + "\x1b[0m"
	snapshot, err := Pager(context.Background(), []byte(colored))
	if err != nil {
		t.Fatal(err)
	}
	assertNames(t, snapshot, "file.txt")
	if snapshot.Repository != root || len(snapshot.Arguments) != 0 {
		t.Fatalf("snapshot metadata: %#v", snapshot)
	}
	if _, err := time.Parse(time.RFC3339, snapshot.CreatedAt); err != nil {
		t.Fatal(err)
	}
	file := snapshot.Files[0]
	if file.Patch != normalizePagerPrefixes(raw) {
		t.Fatalf("patch changed: %q", file.Patch)
	}
	want := normalizePagerPrefixes(blobPatch(oldID, newID, "@@ -1,4 +1,4 @@\n before\n-old\n+new\n after\n last\n\\ No newline at end of file\n"))
	if file.FullPatch != want {
		t.Fatalf("full patch %q; want %q", file.FullPatch, want)
	}
	writeTest(t, "file.txt", "changed after snapshot\n")
	if snapshot.Files[0].FullPatch != want {
		t.Fatal("snapshot changed after editing working tree")
	}
}

func TestPagerRejectsUnsupportedPatchesAndCancellation(t *testing.T) {
	for _, patch := range []string{"", "diff --cc file.txt\n", "diff --git a/file.txt b/file.txt\n@@ -1 +1 @@\n!word diff\n"} {
		if snapshot, err := Pager(context.Background(), []byte(patch)); err == nil || snapshot != nil {
			t.Errorf("accepted %q: %#v %v", patch, snapshot, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if snapshot, err := Pager(ctx, []byte("diff --git a/file.txt b/file.txt\nold mode 100644\nnew mode 100755\n")); !errors.Is(err, context.Canceled) || snapshot != nil {
		t.Fatalf("canceled: %#v %v", snapshot, err)
	}
}

func TestReconstructContext(t *testing.T) {
	testRepo(t)
	old := "before\nold\nafter\n"
	newContent := "before\nnew\nafter\n"
	oldID, newID := storeTestBlob(t, old), storeTestBlob(t, newContent)
	body := "@@ -2 +2 @@\n-old\n+new\n"
	// These IDs are valid hashes, but only one side is stored in the repository.
	unstoredOld := strings.TrimSpace(gitHash(t, "before\nunstored old\nafter\n"))
	unstoredNew := strings.TrimSpace(gitHash(t, "before\nunstored new\nafter\n"))
	for _, tt := range []struct {
		name, kind, patch string
		full              bool
		content           *string
	}{
		{name: "both blobs", patch: blobPatch(oldID, newID, body), full: true},
		{name: "forward with unstored new", patch: blobPatch(oldID, unstoredNew, "@@ -2 +2 @@\n-old\n+unstored new\n"), full: true},
		{name: "reverse with unstored old", patch: blobPatch(unstoredOld, newID, "@@ -2 +2 @@\n-unstored old\n+new\n"), full: true},
		{name: "zero working tree hash", patch: blobPatch(oldID, "0000000", body), full: true},
		{name: "metadata", patch: blobPatch(oldID, oldID, ""), content: &old},
		{name: "different metadata blobs", patch: blobPatch(oldID, newID, "")},
		{name: "binary", patch: blobPatch(oldID, newID, "Binary files i/file.txt and w/file.txt differ\n")},
		{name: "binary patch", patch: blobPatch(oldID, newID, "GIT binary patch\n")},
		{name: "malformed hunks", patch: blobPatch(oldID, newID, "@@ invalid\n")},
		{name: "no index", patch: "diff --git a/file.txt b/file.txt\n"},
		{name: "missing blobs", patch: blobPatch("deadbeef", "cafebabe", body)},
		{name: "wrong context", patch: blobPatch(oldID, newID, "@@ -2 +2 @@\n-wrong\n+new\n")},
		{name: "wrong stored target", patch: blobPatch(oldID, newID, "@@ -2 +2 @@\n-old\n+wrong\n")},
		{name: "wrong target hash", patch: blobPatch(oldID, unstoredNew, body)},
		{name: "wrong reverse context", patch: blobPatch(unstoredOld, newID, "@@ -2 +2 @@\n-unstored old\n+wrong\n")},
		{name: "wrong reverse hash", patch: blobPatch(unstoredOld, newID, body)},
		{name: "already complete", patch: blobPatch(oldID, newID, "@@ -1,3 +1,3 @@\n before\n-old\n+new\n after\n")},
		{name: "added", kind: "added", patch: blobPatch("0000000", newID, "@@ -0,0 +1,3 @@\n+before\n+new\n+after\n")},
		{name: "deleted", kind: "deleted", patch: blobPatch(oldID, "0000000", "@@ -1,3 +0,0 @@\n-before\n-old\n-after\n")},
		{name: "delete all with zero working hash", patch: blobPatch(oldID, "0000000", "@@ -2,2 +1,0 @@\n-old\n-after\n"), full: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			file := File{Kind: tt.kind, Patch: tt.patch}
			reconstructContext(context.Background(), &file)
			if (file.FullPatch != "") != tt.full {
				t.Fatalf("full patch %q; expected expanded=%v", file.FullPatch, tt.full)
			}
			if tt.full {
				_, hunks, err := parseHunks(file.FullPatch)
				if err != nil || len(hunks) != 1 {
					t.Fatalf("invalid expanded patch: %q %v", file.FullPatch, err)
				}
			}
			if tt.content == nil {
				if file.FullContent != nil {
					t.Fatalf("unexpected content %q", *file.FullContent)
				}
			} else if file.FullContent == nil || *file.FullContent != *tt.content {
				t.Fatalf("content %v; want %q", file.FullContent, *tt.content)
			}
		})
	}
}

func gitHash(t *testing.T, content string) string {
	t.Helper()
	cmd := Command(context.Background(), "hash-object", "--stdin")
	cmd.Stdin = strings.NewReader(content)
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestBlobValidation(t *testing.T) {
	testRepo(t)
	id := storeTestBlob(t, "text\n")
	binary := storeTestBlob(t, "binary\x00content")
	for _, tt := range []struct {
		id, want string
		ok       bool
	}{{id, "text\n", true}, {"0000000", "", false}, {"deadbeef", "", false}, {binary, "binary\x00content", false}} {
		got, ok := readBlob(context.Background(), tt.id)
		if got != tt.want || ok != tt.ok {
			t.Errorf("read %s: %q %v", tt.id, got, ok)
		}
	}
	for _, tt := range []struct {
		content, id string
		want        bool
	}{{"text\n", id, true}, {"text\n", id[:7], true}, {"wrong", id, false}, {"anything", "0000000", true}} {
		if got := matchesBlob(context.Background(), tt.content, tt.id); got != tt.want {
			t.Errorf("match %q %s: %v", tt.content, tt.id, got)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if matchesBlob(ctx, "text\n", id) {
		t.Fatal("matched blob with canceled context")
	}
	// A reconstructed text patch must never expose binary content as full context.
	file := File{Patch: blobPatch(binary, "0000000", "@@ -1 +1 @@\n-binary\x00content\n\\ No newline at end of file\n+text\n")}
	reconstructContext(context.Background(), &file)
	if file.FullPatch != "" || file.FullContent != nil {
		t.Fatalf("binary reconstruction: %#v", file)
	}
}
