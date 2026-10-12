package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"diff-punk/cli/internal/git"
	"diff-punk/cli/internal/server"
)

var version = "dev"

// Release builds disable Vite integration through -X main.development=false.
var development = "true"

func main() { os.Exit(run()) }

func run() int {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "diff punk:", err)
		return 1
	}
	if opts.Version {
		fmt.Println("diff punk", version)
		return 0
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if opts.Pager {
		return runPager(ctx)
	}

	args := opts.GitArgs

	if git.TerminalMode(args) {
		return passthrough(ctx, args)
	}

	// Git implicitly uses --no-index outside a repository. Keep that comparison
	// and its exit status in the terminal, just like explicit --no-index.
	if !git.IsGitRepo(ctx) {
		return passthrough(ctx, args)
	}

	snapshot, raw, err := git.Run(ctx, args, os.Stderr)
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
		root, err := git.Command(ctx, "rev-parse", "--show-toplevel").Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "diff punk:", err)
			return 1
		}
		snapshot = &git.Snapshot{
			Repository: strings.TrimSpace(string(root)),
			Arguments:  args,
			CreatedAt:  time.Now().UTC().Format(time.RFC3339),
			Files:      []git.File{},
		}
	}
	repository := snapshot.Repository
	refresh := func(ctx context.Context) (*git.Snapshot, error) {
		var diagnostics bytes.Buffer
		next, raw, err := git.Run(ctx, args, &diagnostics)
		if err != nil {
			if diagnostics.Len() > 0 {
				err = fmt.Errorf("%w: %s", err, strings.TrimSpace(diagnostics.String()))
			}
			return nil, err
		}
		if raw != "" {
			return nil, fmt.Errorf("comparison produced an unsupported patch format")
		}
		if next == nil {
			next = &git.Snapshot{
				Repository: repository,
				Arguments:  args,
				CreatedAt:  time.Now().UTC().Format(time.RFC3339),
				Files:      []git.File{},
			}
		}
		return next, nil
	}

	if err := server.Serve(ctx, snapshot, development == "true", refresh); err != nil {
		fmt.Fprintln(os.Stderr, "diff punk:", err)
		return 1
	}
	return 0
}
