package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"diff-punk/cli/internal/git"
	"diff-punk/cli/internal/server"
)

var version = "dev"

// Release builds disable Vite integration through -X main.development=false.
var development = "true"

func main() { os.Exit(run()) }

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

	if git.TerminalMode(args) {
		return passthrough(ctx, args)
	}

	// Git implicitly uses --no-index outside a repository. Keep that comparison
	// and its exit status in the terminal, just like explicit --no-index.
	if err := git.Command(ctx, "rev-parse", "--git-dir").Run(); err != nil {
		return passthrough(ctx, args)
	}

	snapshot, raw, err := git.Capture(ctx, os.Args[1:])
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
		return 0
	}
	if err := server.Serve(ctx, snapshot, development == "true"); err != nil {
		fmt.Fprintln(os.Stderr, "diff punk:", err)
		return 1
	}
	return 0
}

func passthrough(ctx context.Context, args []string) int {
	cmd := git.Command(ctx, append([]string{"diff"}, args...)...)
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
