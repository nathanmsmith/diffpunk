package git

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseHunks(t *testing.T) {
	for _, tt := range []struct {
		name, patch, prefix string
		hunks               []patchHunk
	}{
		{name: "metadata", patch: "old mode 100644\nnew mode 100755\n", prefix: "old mode 100644\nnew mode 100755\n", hunks: []patchHunk{}},
		{name: "implicit counts and context", patch: "header\n@@ -2 +2 @@ function\n-old\n+new\n", prefix: "header\n", hunks: []patchHunk{{2, 1, 2, 1, []patchLine{{'-', "old\n"}, {'+', "new\n"}}}}},
		{name: "multiple hunks and empty ranges", patch: "@@ -0,0 +1,1 @@\n+first\n@@ -2,1 +2,0 @@\n-last\n", hunks: []patchHunk{{0, 0, 1, 1, []patchLine{{'+', "first\n"}}}, {2, 1, 2, 0, []patchLine{{'-', "last\n"}}}}},
		{name: "no newline", patch: "@@ -1,2 +1,2 @@\n context\n-old\n\\ No newline at end of file\n+new\n\\ No newline at end of file", hunks: []patchHunk{{1, 2, 1, 2, []patchLine{{' ', "context\n"}, {'-', "old"}, {'+', "new"}}}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			prefix, hunks, err := parseHunks(tt.patch)
			if err != nil || prefix != tt.prefix || !reflect.DeepEqual(hunks, tt.hunks) {
				t.Fatalf("got %q %#v %v; want %q %#v", prefix, hunks, err, tt.prefix, tt.hunks)
			}
		})
	}
}

func TestParseHunksRejectsMalformedPatches(t *testing.T) {
	for name, patch := range map[string]string{
		"combined": "@@@ -1 -1 +1 @@@\n", "overflow": "@@ -" + strings.Repeat("9", 100) + " +1 @@\n-a\n+b\n",
		"marker before lines": "@@ -1 +1 @@\n\\ No newline at end of file\n", "duplicate marker": "@@ -1 +1 @@\n same\n\\ No newline at end of file\n\\ No newline at end of file\n",
		"unknown body": "@@ -1 +1 @@\n!word diff\n", "wrong counts": "@@ -1,2 +1 @@\n-old\n+new\n", "zero old start": "@@ -0 +1 @@\n-old\n+new\n", "zero new start": "@@ -1 +0 @@\n-old\n+new\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := parseHunks(patch); err == nil {
				t.Fatal("accepted malformed patch")
			}
		})
	}
}

func TestApplyHunks(t *testing.T) {
	for _, tt := range []struct{ name, source, patch, want string }{
		{"replace", "before\nold\nafter\n", "@@ -2 +2 @@\n-old\n+new\n", "before\nnew\nafter\n"},
		{"insert at beginning", "tail\n", "@@ -0,0 +1 @@\n+first\n", "first\ntail\n"},
		{"delete", "first\ntail\n", "@@ -1 +0,0 @@\n-first\n", "tail\n"},
		{"empty source", "", "@@ -0,0 +1 @@\n+first\n", "first\n"},
		{"unterminated line", "old", "@@ -1 +1 @@\n-old\n\\ No newline at end of file\n+new\n\\ No newline at end of file\n", "new"},
		{"multiple hunks", "a\nb\nc\nd\ne\n", "@@ -2 +2 @@\n-b\n+B\n@@ -4 +4 @@\n-d\n+D\n", "a\nB\nc\nD\ne\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, hunks, err := parseHunks(tt.patch)
			if err != nil {
				t.Fatal(err)
			}
			got, ok := applyHunks(tt.source, hunks, false)
			if !ok || got != tt.want {
				t.Fatalf("forward %q %v; want %q", got, ok, tt.want)
			}
			got, ok = applyHunks(tt.want, hunks, true)
			if !ok || got != tt.source {
				t.Fatalf("reverse %q %v; want %q", got, ok, tt.source)
			}
		})
	}
}

func TestApplyHunksRejectsInvalidContext(t *testing.T) {
	for name, patch := range map[string]string{
		"out of bounds": "@@ -4 +4 @@\n-old\n+new\n", "too many source lines": "@@ -2,2 +2 @@\n-old\n-tail\n+new\n", "target offset": "@@ -1 +2 @@\n-old\n+new\n", "context mismatch": "@@ -1 +1 @@\n-other\n+new\n", "overlapping hunks": "@@ -1 +1 @@\n-old\n+new\n@@ -1 +1 @@\n-old\n+new\n", "interior unterminated line": "@@ -1 +1 @@\n-old\n+new\n\\ No newline at end of file\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, hunks, err := parseHunks(patch)
			if err != nil {
				t.Fatal(err)
			}
			if got, ok := applyHunks("old\ntail\n", hunks, false); ok {
				t.Fatalf("accepted invalid context: %q", got)
			}
		})
	}
	// An insertion cannot extend an unterminated source line into the middle of a file.
	if got, ok := applyHunks("old", []patchHunk{{1, 0, 2, 1, []patchLine{{'+', "new\n"}}}}, false); ok {
		t.Fatalf("accepted interior unterminated source: %q", got)
	}
	if got, ok := applyHunks("", []patchHunk{{0, 0, 0, 0, []patchLine{{' ', "missing\n"}}}}, false); ok {
		t.Fatalf("read past source: %q", got)
	}
}

func TestSplitContentAndWritePatchLine(t *testing.T) {
	for _, tt := range []struct {
		content string
		want    []string
	}{{"", nil}, {"a\nb\n", []string{"a\n", "b\n"}}, {"a\nb", []string{"a\n", "b"}}} {
		if got := splitContent(tt.content); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("split %q: %v; want %v", tt.content, got, tt.want)
		}
	}
	var out strings.Builder
	writePatchLine(&out, patchLine{' ', "context\n"})
	writePatchLine(&out, patchLine{'+', "last"})
	if got, want := out.String(), " context\n+last\n\\ No newline at end of file\n"; got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}
