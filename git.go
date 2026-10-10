package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type File struct {
	Name         string  `json:"name"`
	PreviousName string  `json:"previousName,omitempty"`
	Kind         string  `json:"kind"`
	Patch        string  `json:"patch"`
	FullPatch    string  `json:"fullPatch,omitempty"`
	FullContent  *string `json:"fullContent,omitempty"`
}

type Snapshot struct {
	Repository         string   `json:"repository"`
	Arguments          []string `json:"arguments"`
	CreatedAt          string   `json:"createdAt"`
	Files              []File   `json:"files"`
	UntrackedAvailable bool     `json:"untrackedAvailable"`
	ShowUntracked      bool     `json:"showUntracked"`
}

// Reporting and alternative formats retain their native terminal behavior.
func terminalMode(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			break
		}
		name := strings.SplitN(arg, "=", 2)[0]
		switch name {
		case "--stat", "--numstat", "--shortstat", "--dirstat", "--dirstat-by-file", "--cumulative", "--summary", "--compact-summary", "--name-only", "--name-status", "--check", "--raw", "--patch-with-raw", "--patch-with-stat", "--word-diff", "--word-diff-regex", "--color-words", "--no-patch", "--quiet", "--exit-code", "--help", "-h", "-s", "--output", "--binary", "--cc", "-c", "--combined-all-paths", "--ext-diff", "--no-index", "--submodule", "--line-prefix", "--src-prefix", "--dst-prefix", "--no-prefix", "--default-prefix", "--output-indicator-new", "--output-indicator-old", "--output-indicator-context", "--irreversible-delete", "-D":
			return true
		}
		if strings.HasPrefix(name, "--stat-") || strings.HasPrefix(name, "--dirstat=") {
			return true
		}
	}
	return false
}

func gitCommand(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "git", append([]string{"--no-pager"}, args...)...)
}

func passthrough(ctx context.Context, args []string) int {
	cmd := gitCommand(ctx, append([]string{"diff"}, args...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return e.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "diff punk:", err)
		return 1
	}
	return 0
}

// Insert format controls before the path separator, leaving all Git selection
// options, revisions, and pathspecs intact.
func patchArgs(args []string, full bool) []string {
	controls := []string{"--no-color", "--src-prefix=a/", "--dst-prefix=b/"}
	if full {
		controls = append(controls, "--unified=2147483647")
	}
	return diffArgs(args, controls)
}

func diffArgs(args, controls []string) []string {
	result := []string{"-c", "diff.noprefix=false", "-c", "diff.mnemonicPrefix=false", "diff"}
	for i, arg := range args {
		if arg == "--" {
			result = append(result, controls...)
			return append(result, args[i:]...)
		}
		result = append(result, arg)
	}
	return append(result, controls...)
}

func capture(ctx context.Context, args []string) (*Snapshot, string, error) {
	args, showUntracked := viewerArgs(args)
	cmd := gitCommand(ctx, patchArgs(args, false)...)
	cmd.Stderr = os.Stderr
	patch, err := cmd.Output()
	if err != nil {
		return nil, "", err
	}
	files := []File{}
	if len(patch) > 0 {
		files, err = parseFiles(string(patch))
		if err != nil {
			return nil, string(patch), nil
		}
	}
	// A second patch uses the same selection and diff algorithm, but includes all
	// context. This also works for staged and historical files and textconv.
	fullCmd := gitCommand(ctx, patchArgs(args, true)...)
	var fullErr bytes.Buffer
	fullCmd.Stderr = &fullErr
	full, err := fullCmd.Output()
	if err == nil {
		expanded, parseErr := parseFiles(string(full))
		if parseErr == nil {
			byName := make(map[string]string, len(expanded))
			for _, f := range expanded {
				byName[f.Name] = f.Patch
			}
			for i := range files {
				files[i].FullPatch = byName[files[i].Name]
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "diff punk: full-file context unavailable:", strings.TrimSpace(fullErr.String()))
	}
	captureMetadataContents(ctx, args, files)
	rootCmd := gitCommand(ctx, "rev-parse", "--show-toplevel")
	root, _ := rootCmd.Output()
	available, paths, relative := untrackedSelection(ctx, args)
	if available {
		untracked, captureErr := captureUntracked(ctx, strings.TrimSpace(string(root)), paths, relative)
		if captureErr != nil {
			return nil, "", captureErr
		}
		files = append(files, untracked...)
	}
	if len(files) == 0 {
		return nil, "", nil
	}
	return &Snapshot{
		Repository: strings.TrimSpace(string(root)), Arguments: append([]string{}, args...),
		CreatedAt: time.Now().UTC().Format(time.RFC3339), Files: sortFiles(files),
		UntrackedAvailable: available, ShowUntracked: showUntracked && available,
	}, "", nil
}

var fileStart = regexp.MustCompile(`(?m)^diff --git `)

func parseFiles(patch string) ([]File, error) {
	starts := fileStart.FindAllStringIndex(patch, -1)
	if len(starts) == 0 || starts[0][0] != 0 || strings.Contains(patch, "\ndiff --cc ") || strings.Contains(patch, "\ndiff --combined ") {
		return nil, fmt.Errorf("unsupported patch format")
	}
	files := make([]File, 0, len(starts))
	for i, start := range starts {
		end := len(patch)
		if i+1 < len(starts) {
			end = starts[i+1][0]
		}
		raw := patch[start[0]:end]
		lines := strings.Split(raw, "\n")
		oldName, name, err := headerNames(strings.TrimPrefix(lines[0], "diff --git "))
		if err != nil {
			return nil, err
		}
		f := File{Name: name, PreviousName: oldName, Kind: "modified", Patch: raw}
		for _, line := range lines[1:] {
			if strings.HasPrefix(line, "@@") {
				break
			}
			switch {
			case strings.HasPrefix(line, "--- "):
				if n := patchName(line[4:], true); n != "/dev/null" {
					f.PreviousName = n
				}
			case strings.HasPrefix(line, "+++ "):
				if n := patchName(line[4:], true); n != "/dev/null" {
					f.Name = n
				}
			case strings.HasPrefix(line, "new file mode "):
				f.Kind = "added"
			case strings.HasPrefix(line, "deleted file mode "):
				f.Kind = "deleted"
			case strings.HasPrefix(line, "rename from "):
				f.Kind = "renamed"
				f.PreviousName = patchName(line[len("rename from "):], false)
			case strings.HasPrefix(line, "rename to "):
				f.Name = patchName(line[len("rename to "):], false)
			case strings.HasPrefix(line, "copy from "):
				f.Kind = "copied"
				f.PreviousName = patchName(line[len("copy from "):], false)
			case strings.HasPrefix(line, "copy to "):
				f.Name = patchName(line[len("copy to "):], false)
			}
		}
		if f.PreviousName == f.Name {
			f.PreviousName = ""
		}
		files = append(files, f)
	}
	return files, nil
}

func patchName(s string, prefixed bool) string {
	s = strings.TrimSuffix(s, "\r")
	if strings.HasPrefix(s, `"`) {
		if decoded, err := strconv.Unquote(s); err == nil {
			s = decoded
		}
	} else {
		s = strings.TrimSuffix(s, "\t")
	}
	if prefixed && (strings.HasPrefix(s, "a/") || strings.HasPrefix(s, "b/")) {
		s = s[2:]
	}
	return s
}

func headerNames(s string) (string, string, error) {
	var old, rest string
	if strings.HasPrefix(s, `"`) {
		escaped := false
		for i := 1; i < len(s); i++ {
			if s[i] == '"' && !escaped {
				old, rest = s[:i+1], strings.TrimPrefix(s[i+1:], " ")
				break
			}
			if s[i] == '\\' {
				escaped = !escaped
			} else {
				escaped = false
			}
		}
	} else {
		// Git leaves spaces unquoted; text headers and rename metadata below
		// provide the definitive names for those files.
		split := strings.Index(s, " b/")
		if split == -1 {
			split = strings.Index(s, ` "b/`)
		}
		if split >= 0 {
			old, rest = s[:split], s[split+1:]
			// For ordinary changes, both paths are identical. Check all
			// delimiters because a filename itself may contain " b/".
			for offset := 0; offset < len(s); {
				next := strings.Index(s[offset:], " b/")
				if next < 0 {
					break
				}
				next += offset
				left, right := s[:next], s[next+1:]
				if patchName(left, true) == patchName(right, true) {
					old, rest = left, right
					break
				}
				offset = next + 1
			}
		}
	}
	if old == "" || rest == "" {
		return "", "", fmt.Errorf("invalid diff header: %s", s)
	}
	return patchName(old, true), patchName(rest, true), nil
}

// Pure renames and permission changes have no patch hunks. Their original
// blobs still let the user inspect the whole unchanged text file.
func captureMetadataContents(ctx context.Context, args []string, files []File) {
	wanted := map[string]int{}
	for i, f := range files {
		if !strings.Contains(f.Patch, "\n@@ ") && !strings.Contains(f.Patch, "Binary files ") && !strings.Contains(f.Patch, "GIT binary patch") {
			wanted[f.Name] = i
		}
	}
	if len(wanted) == 0 {
		return
	}
	raw, err := gitCommand(ctx, diffArgs(args, []string{"--no-color", "--no-patch", "--raw", "-z", "--no-abbrev"})...).Output()
	if err != nil {
		return
	}
	records := strings.Split(string(raw), "\x00")
	for i := 0; i+1 < len(records); {
		fields := strings.Fields(records[i])
		if len(fields) != 5 || !strings.HasPrefix(fields[0], ":") {
			break
		}
		name := records[i+1]
		i += 2
		if fields[4][0] == 'R' || fields[4][0] == 'C' {
			if i >= len(records) {
				break
			}
			name = records[i]
			i++
		}
		index, ok := wanted[name]
		if !ok || fields[0] == ":160000" || fields[1] == "160000" {
			continue
		}
		object := fields[2]
		if strings.Trim(object, "0") == "" {
			object = fields[3]
		}
		var content []byte
		if strings.Trim(object, "0") != "" {
			content, err = gitCommand(ctx, "cat-file", "blob", object).Output()
			if err != nil || bytes.IndexByte(content, 0) >= 0 {
				continue
			}
		}
		text := string(content)
		files[index].FullContent = &text
	}
}
