package main

import (
	"context"
	"diff-punk/cli/internal/git"
	"diff-punk/cli/internal/server"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func run() int {
	args := os.Args[1:]
	if len(args) == 1 && args[0] == "--version" {
		fmt.Println("diff punk", version)
		return 0
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if len(args) > 0 && args[0] == "--pager" {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "diff punk: --pager reads Git output from stdin and takes no diff arguments")
			return 1
		}
		return runPager(ctx)
	}
	watch := true
	filtered := []string{}
	paths := false
	for _, arg := range args {
		if arg == "--" {
			paths = true
		}
		if !paths && arg == "--no-watch" {
			watch = false
			continue
		}
		filtered = append(filtered, arg)
	}
	args = filtered

	if git.TerminalMode(args) {
		return passthrough(ctx, args)
	}

	// Git implicitly uses --no-index outside a repository. Keep that comparison
	// and its exit status in the terminal, just like explicit --no-index.
	if err := git.Command(ctx, "rev-parse", "--git-dir").Run(); err != nil {
		return passthrough(ctx, args)
	}

	snapshot, raw, err := git.Capture(ctx, args)
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return e.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "diff punk:", err)
		return 1
	}
	if raw != "" {
		fmt.Print(raw)
		return 0
	}
	if snapshot == nil {
		if !watch {
			return 0
		}
		root, err := git.Command(ctx, "rev-parse", "--show-toplevel").Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "diff punk:", err)
			return 1
		}
		snapshot = &git.Snapshot{Repository: strings.TrimSpace(string(root)), Arguments: args,
			CreatedAt: time.Now().UTC().Format(time.RFC3339), Files: []git.File{}}
	}
	var refresh server.Refresh
	if watch {
		repository := snapshot.Repository
		refresh = func(ctx context.Context) (*git.Snapshot, error) {
			next, raw, err := git.CaptureRefresh(ctx, args)
			if err != nil {
				return nil, err
			}
			if raw != "" {
				return nil, fmt.Errorf("comparison produced an unsupported patch format")
			}
			if next == nil {
				next = &git.Snapshot{Repository: repository, Arguments: args,
					CreatedAt: time.Now().UTC().Format(time.RFC3339), Files: []git.File{}}
			}
			return next, nil
		}
	}
	if err := server.Serve(ctx, snapshot, development == "true", refresh); err != nil {
		fmt.Fprintln(os.Stderr, "diff punk:", err)
		return 1
	}
	return 0
}
