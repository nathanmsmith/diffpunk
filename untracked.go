package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Viewer flags never reach Git; arguments after -- are always pathspecs.
func viewerArgs(args []string) ([]string, bool) {
	result := []string{}
	show := false
	for i, arg := range args {
		if arg == "--" {
			return append(result, args[i:]...), show
		}
		if arg == "--untracked" {
			show = true
		} else {
			result = append(result, arg)
		}
	}
	return result, show
}

// Let Git resolve positional arguments into revisions and pathspecs. Diff
// options with separate values must first be removed so their values cannot
// accidentally be interpreted as paths or revisions.
func untrackedSelection(ctx context.Context, args []string) (bool, []string, string) {
	positional := []string{}
	paths := []string{}
	relative := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			paths = append(paths, args[i+1:]...)
			break
		}
		if arg == "--cached" || arg == "--staged" {
			return false, nil, ""
		}
		if arg == "--relative" {
			prefix, err := gitCommand(ctx, "rev-parse", "--show-prefix").Output()
			if err != nil {
				return false, nil, ""
			}
			relative = strings.TrimSpace(string(prefix))
		} else if strings.HasPrefix(arg, "--relative=") {
			relative = strings.TrimPrefix(arg, "--relative=")
			if relative != "" {
				relative = strings.TrimSuffix(relative, "/") + "/"
			}
		} else if arg == "--no-relative" {
			relative = ""
		}
		if strings.HasPrefix(arg, "-") {
			switch arg {
			case "-U", "--unified", "--inter-hunk-context", "-S", "-G", "-O", "-l", "--diff-filter", "--diff-algorithm", "--anchored", "--find-object", "-I", "--ignore-matching-lines", "--rotate-to", "--skip-to":
				i++
			}
			continue
		}
		positional = append(positional, arg)
	}
	if len(positional) > 0 {
		revs, err := gitCommand(ctx, append([]string{"rev-parse", "--revs-only", "--no-flags"}, positional...)...).Output()
		if err != nil || len(strings.Fields(string(revs))) > 1 {
			return false, nil, ""
		}
		// Ask Git one argument at a time to preserve filenames containing newlines.
		for _, arg := range positional {
			rev, err := gitCommand(ctx, "rev-parse", "--revs-only", "--no-flags", arg).Output()
			if err != nil {
				return false, nil, ""
			}
			if len(rev) == 0 {
				paths = append(paths, arg)
			}
		}
	}
	return true, paths, relative
}

func captureUntracked(ctx context.Context, root string, paths []string, relative string) ([]File, error) {
	args := []string{"ls-files", "--others", "--exclude-standard", "--full-name", "-z", "--"}
	names, err := gitCommand(ctx, append(args, paths...)...).Output()
	if err != nil {
		return nil, fmt.Errorf("list untracked files: %w", err)
	}
	files := []File{}
	for _, name := range strings.Split(string(names), "\x00") {
		if name == "" || !strings.HasPrefix(name, relative) {
			continue
		}
		info, err := os.Lstat(filepath.Join(root, name))
		if err != nil {
			return nil, fmt.Errorf("inspect untracked file %q: %w", name, err)
		}
		// Git lists embedded repositories as directories. They have no file patch.
		if info.IsDir() {
			continue
		}
		cmd := gitCommand(ctx, "-c", "diff.noprefix=false", "-c", "diff.mnemonicPrefix=false", "diff", "--no-index", "--no-ext-diff", "--no-textconv", "--no-color", "--src-prefix=a/", "--dst-prefix=b/", "--", "/dev/null", name)
		cmd.Dir = root
		patch, err := cmd.Output()
		if err != nil {
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 {
				return nil, fmt.Errorf("capture untracked file %q: %w", name, err)
			}
		}
		parsed, err := parseFiles(string(patch))
		if err != nil {
			return nil, fmt.Errorf("parse untracked file %q: %w", name, err)
		}
		for _, file := range parsed {
			file.Name = strings.TrimPrefix(name, relative)
			file.PreviousName = ""
			file.Kind = "untracked"
			files = append(files, file)
		}
	}
	return files, nil
}
