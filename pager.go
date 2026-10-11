package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

// Pager mode consumes Git's output, never reruns an inferred comparison, and
// retains the original bytes for reporting formats and unsupported patches.
func runPager(ctx context.Context) int {
	stopRead := context.AfterFunc(ctx, func() { _ = os.Stdin.Close() })
	defer stopRead()
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		_, err := io.Copy(os.Stdout, os.Stdin)
		if ctx.Err() != nil {
			return 0
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "diff punk:", err)
			return 1
		}
		return 0
	}
	raw, err := io.ReadAll(os.Stdin)
	if ctx.Err() != nil {
		return 0
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "diff punk:", err)
		return 1
	}
	if len(raw) == 0 {
		return 0
	}
	files, err := parseFiles(terminalColor.ReplaceAllString(string(raw), ""))
	if err == nil {
		for _, file := range files {
			if _, _, parseErr := parseHunks(file.Patch); parseErr != nil {
				err = parseErr
				break
			}
		}
	}
	if err != nil {
		if _, err := os.Stdout.Write(raw); err != nil {
			fmt.Fprintln(os.Stderr, "diff punk:", err)
			return 1
		}
		return 0
	}
	for i := range files {
		reconstructContext(ctx, &files[i])
	}
	if ctx.Err() != nil {
		return 0
	}
	root, _ := gitCommand(ctx, "rev-parse", "--show-toplevel").Output()
	snapshot := &Snapshot{
		Repository: strings.TrimSpace(string(root)), Arguments: []string{},
		CreatedAt: time.Now().UTC().Format(time.RFC3339), Files: sortFiles(files),
	}
	if err := serve(ctx, snapshot); err != nil {
		fmt.Fprintln(os.Stderr, "diff punk:", err)
		return 1
	}
	return 0
}

var (
	terminalColor = regexp.MustCompile(`\x1b\[[0-9;:]*m`)
	blobHeader    = regexp.MustCompile(`(?m)^index ([0-9a-f]{4,64})\.\.([0-9a-f]{4,64})(?: [0-7]+)?$`)
	hunkHeader    = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@[^\n]*\n?$`)
)

type patchLine struct {
	kind byte
	text string // Includes the newline, unless Git supplied a no-newline marker.
}

type patchHunk struct {
	oldStart, oldCount int
	newStart, newCount int
	lines              []patchLine
}

// parseHunks rejects word/combined diffs and malformed counts before the browser
// is opened. Binary and metadata-only changes have no text hunks.
func parseHunks(patch string) (string, []patchHunk, error) {
	lines := strings.SplitAfter(patch, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	prefix := ""
	hunks := []patchHunk{}
	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			match := hunkHeader.FindStringSubmatch(line)
			if match == nil {
				return "", nil, fmt.Errorf("unsupported hunk header")
			}
			values := [4]int{}
			for i := range values {
				if match[i+1] == "" {
					values[i] = 1
					continue
				}
				n, err := strconv.Atoi(match[i+1])
				if err != nil {
					return "", nil, err
				}
				values[i] = n
			}
			hunks = append(hunks, patchHunk{oldStart: values[0], oldCount: values[1], newStart: values[2], newCount: values[3]})
			continue
		}
		if len(hunks) == 0 {
			prefix += line
			continue
		}
		hunk := &hunks[len(hunks)-1]
		if line == "\\ No newline at end of file\n" || line == "\\ No newline at end of file" {
			if len(hunk.lines) == 0 || !strings.HasSuffix(hunk.lines[len(hunk.lines)-1].text, "\n") {
				return "", nil, fmt.Errorf("invalid no-newline marker")
			}
			hunk.lines[len(hunk.lines)-1].text = strings.TrimSuffix(hunk.lines[len(hunk.lines)-1].text, "\n")
			continue
		}
		if len(line) == 0 || (line[0] != ' ' && line[0] != '-' && line[0] != '+') {
			return "", nil, fmt.Errorf("unsupported patch body")
		}
		hunk.lines = append(hunk.lines, patchLine{kind: line[0], text: line[1:]})
	}
	for _, hunk := range hunks {
		oldCount, newCount := 0, 0
		for _, line := range hunk.lines {
			if line.kind != '+' {
				oldCount++
			}
			if line.kind != '-' {
				newCount++
			}
		}
		if oldCount != hunk.oldCount || newCount != hunk.newCount ||
			(hunk.oldCount > 0 && hunk.oldStart == 0) || (hunk.newCount > 0 && hunk.newStart == 0) {
			return "", nil, fmt.Errorf("invalid hunk counts")
		}
	}
	return prefix, hunks, nil
}

func hunkOffset(start, count int) int {
	if count == 0 {
		return start
	}
	return start - 1
}

// applyHunks requires exact positions and byte-for-byte matching context. Reverse
// application allows reconstruction when only the new blob is available.
func applyHunks(source string, hunks []patchHunk, reverse bool) (string, bool) {
	lines := splitContent(source)
	result := []string{}
	cursor := 0
	for _, hunk := range hunks {
		start, count := hunk.oldStart, hunk.oldCount
		targetStart, targetCount := hunk.newStart, hunk.newCount
		remove, insert := byte('-'), byte('+')
		if reverse {
			start, count, targetStart, targetCount = targetStart, targetCount, start, count
			remove, insert = insert, remove
		}
		offset := hunkOffset(start, count)
		if offset < cursor || offset > len(lines) || count > len(lines)-offset {
			return "", false
		}
		result = append(result, lines[cursor:offset]...)
		if hunkOffset(targetStart, targetCount) != len(result) {
			return "", false
		}
		cursor = offset
		for _, line := range hunk.lines {
			if line.kind != insert {
				if cursor >= len(lines) || lines[cursor] != line.text {
					return "", false
				}
				cursor++
			}
			if line.kind != remove {
				result = append(result, line.text)
			}
		}
	}
	result = append(result, lines[cursor:]...)
	// A line without a newline can only occur at the end of a text file.
	for _, line := range result[:max(0, len(result)-1)] {
		if !strings.HasSuffix(line, "\n") {
			return "", false
		}
	}
	return strings.Join(result, ""), true
}

func splitContent(content string) []string {
	if content == "" {
		return nil
	}
	lines := strings.SplitAfter(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func readBlob(ctx context.Context, id string) (string, bool) {
	if strings.Trim(id, "0") == "" {
		return "", false
	}
	data, err := gitCommand(ctx, "cat-file", "blob", id).Output()
	return string(data), err == nil && !strings.ContainsRune(string(data), 0)
}

func matchesBlob(ctx context.Context, content, id string) bool {
	if strings.Trim(id, "0") == "" {
		return true // Git sometimes emits zero for an unstored working-tree side.
	}
	cmd := gitCommand(ctx, "hash-object", "--stdin")
	cmd.Stdin = strings.NewReader(content)
	hash, err := cmd.Output()
	return err == nil && strings.HasPrefix(strings.TrimSpace(string(hash)), id)
}

func reconstructContext(ctx context.Context, file *File) {
	if strings.Contains(file.Patch, "\nBinary files ") || strings.Contains(file.Patch, "\nGIT binary patch") {
		return
	}
	prefix, hunks, err := parseHunks(file.Patch)
	if err != nil {
		return
	}
	ids := blobHeader.FindStringSubmatch(prefix)
	if ids == nil {
		return
	}
	old, oldOK := readBlob(ctx, ids[1])
	newContent, newOK := readBlob(ctx, ids[2])
	if file.Kind == "added" && strings.Trim(ids[1], "0") == "" {
		old, oldOK = "", true
	}
	if file.Kind == "deleted" && strings.Trim(ids[2], "0") == "" {
		newContent, newOK = "", true
	}
	if len(hunks) == 0 {
		if oldOK && newOK && old == newContent {
			file.FullContent = &old
		}
		return
	}
	if oldOK {
		built, ok := applyHunks(old, hunks, false)
		if !ok || (newOK && built != newContent) || !matchesBlob(ctx, built, ids[2]) {
			return
		}
		newContent, newOK = built, true
	} else if newOK {
		built, ok := applyHunks(newContent, hunks, true)
		if !ok || !matchesBlob(ctx, built, ids[1]) {
			return
		}
		old, oldOK = built, true
	}
	if !oldOK || !newOK || strings.ContainsRune(old, 0) || strings.ContainsRune(newContent, 0) {
		return
	}
	oldLines, newLines := splitContent(old), splitContent(newContent)
	covered := 0
	for _, hunk := range hunks {
		covered += hunk.oldCount
	}
	if covered == len(oldLines) {
		return // The supplied patch already contains the entire file.
	}
	// Preserve all supplied changed lines. Only add verified unchanged context;
	// recomputing a diff could discard the user's chosen algorithm or filters.
	var full strings.Builder
	full.WriteString(prefix)
	oldStart, newStart := 1, 1
	if len(oldLines) == 0 {
		oldStart = 0
	}
	if len(newLines) == 0 {
		newStart = 0
	}
	fmt.Fprintf(&full, "@@ -%d,%d +%d,%d @@\n", oldStart, len(oldLines), newStart, len(newLines))
	cursor := 0
	for _, hunk := range hunks {
		offset := hunkOffset(hunk.oldStart, hunk.oldCount)
		for _, line := range oldLines[cursor:offset] {
			writePatchLine(&full, patchLine{kind: ' ', text: line})
		}
		for _, line := range hunk.lines {
			writePatchLine(&full, line)
		}
		cursor = offset + hunk.oldCount
	}
	for _, line := range oldLines[cursor:] {
		writePatchLine(&full, patchLine{kind: ' ', text: line})
	}
	file.FullPatch = full.String()
}

func writePatchLine(out *strings.Builder, line patchLine) {
	out.WriteByte(line.kind)
	out.WriteString(line.text)
	if !strings.HasSuffix(line.text, "\n") {
		out.WriteString("\n\\ No newline at end of file\n")
	}
}
