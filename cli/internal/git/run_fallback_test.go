package git

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunReturnsUnsupportedOutputUnchanged(t *testing.T) {
	patch := "diff --cc conflict.txt\n@@@ -1 -1 +1 @@@\n"
	stubGit(t, "cat \"$dir/patch\"\n", map[string]string{"patch": patch})
	var diagnostics bytes.Buffer
	snapshot, raw, err := Run(context.Background(), nil, &diagnostics)
	if snapshot != nil || raw != patch || err != nil {
		t.Fatalf("got %#v %q %v", snapshot, raw, err)
	}
	if diagnostics.Len() != 0 {
		t.Fatalf("unexpected diagnostics: %q", diagnostics.String())
	}
}

func TestRunPreservesSnapshotWhenFullContextUnavailable(t *testing.T) {
	patch := "diff --git a/file b/file\n--- a/file\n+++ b/file\n@@ -1 +1 @@\n-old\n+new\n"
	stubGit(t, `for arg do
 if [ "$arg" = "--unified=2147483647" ]; then
  echo 'full context failed' >&2
  exit 1
 fi
 if [ "$arg" = "rev-parse" ]; then
  for option do
   if [ "$option" = "--show-toplevel" ]; then echo '/repository'; fi
  done
  exit 0
 fi
 if [ "$arg" = "ls-files" ]; then
  exit 0
 fi
 done
 cat "$dir/patch"
`, map[string]string{"patch": patch})
	var diagnostics bytes.Buffer
	args := []string{"HEAD", "--", "file"}
	snapshot, raw, err := Run(context.Background(), args, &diagnostics)
	if err != nil || raw != "" || snapshot == nil {
		t.Fatalf("got %#v %q %v", snapshot, raw, err)
	}
	assertNames(t, snapshot, "file")
	if snapshot.Files[0].Patch != patch || snapshot.Files[0].FullPatch != "" || snapshot.Repository != "/repository" {
		t.Fatalf("snapshot %#v", snapshot)
	}
	if !strings.Contains(diagnostics.String(), "diff punk: full-file context unavailable: full context failed") {
		t.Fatalf("diagnostics %q", diagnostics.String())
	}
	args[0] = "changed"
	if snapshot.Arguments[0] != "HEAD" {
		t.Fatal("snapshot aliases arguments")
	}
}
