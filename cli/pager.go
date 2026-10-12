package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"

	"diff-punk/cli/internal/git"
	"diff-punk/cli/internal/server"
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
	snapshot, err := git.Pager(ctx, raw)
	if ctx.Err() != nil {
		return 0
	}
	if err != nil {
		if _, err := os.Stdout.Write(raw); err != nil {
			fmt.Fprintln(os.Stderr, "diff punk:", err)
			return 1
		}
		return 0
	}
	if err := server.Serve(ctx, snapshot, development == "true"); err != nil {
		fmt.Fprintln(os.Stderr, "diff punk:", err)
		return 1
	}
	return 0
}
