package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// stubGit isolates command failures and malformed raw output that real Git
// normally cannot produce, while still exercising the subprocess boundary.
func stubGit(t *testing.T, script string, fixtures map[string]string) {
	t.Helper()
	dir := t.TempDir()
	for name, content := range fixtures {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte("#!/bin/sh\ndir=${0%/*}\n"+script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestLoadMetadataContentsHandlesUnavailableRawDiff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	files := []File{{Name: "file", Patch: "old mode 100644\nnew mode 100755\n"}}
	loadMetadataContents(ctx, nil, files)
	if files[0].FullContent != nil {
		t.Fatal("loaded content after cancellation")
	}
}

func TestLoadMetadataContentsRawRecords(t *testing.T) {
	for _, tt := range []struct {
		name, raw, content    string
		catFails, wantContent bool
	}{
		{name: "malformed record", raw: "invalid\x00file\x00"},
		{name: "truncated rename", raw: ":100644 100644 abcd abcd R100\x00old"},
		{name: "added empty file", raw: ":000000 100644 0000 abcd A\x00file\x00", content: "", wantContent: true},
		{name: "zero objects", raw: ":000000 100644 0000 0000 A\x00file\x00", wantContent: true},
		{name: "missing object", raw: ":100644 100755 abcd abcd M\x00file\x00", catFails: true},
		{name: "binary object", raw: ":100644 100755 abcd abcd M\x00file\x00", content: "binary\x00data"},
		{name: "unwanted file", raw: ":100644 100755 abcd abcd M\x00other\x00", content: "text\n"},
		{name: "old submodule", raw: ":160000 100644 abcd abcd T\x00file\x00", content: "text\n"},
		{name: "new submodule", raw: ":100644 160000 abcd abcd T\x00file\x00", content: "text\n"},
		{name: "copy", raw: ":100644 100644 abcd abcd C100\x00old\x00file\x00", content: "text\n", wantContent: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			catCommand := "cat \"$dir/content\""
			if tt.catFails {
				catCommand = "exit 1"
			}
			stubGit(t, "case \"$2\" in\n cat-file) "+catCommand+" ;;\n *) cat \"$dir/raw\" ;;\nesac\n", map[string]string{"raw": tt.raw, "content": tt.content})
			files := []File{{Name: "file", Patch: "old mode 100644\nnew mode 100755\n"}}
			loadMetadataContents(context.Background(), nil, files)
			if tt.wantContent {
				if files[0].FullContent == nil || *files[0].FullContent != tt.content {
					t.Fatalf("content %v; want %q", files[0].FullContent, tt.content)
				}
			} else if files[0].FullContent != nil {
				t.Fatalf("unexpected content %q", *files[0].FullContent)
			}
		})
	}
}
