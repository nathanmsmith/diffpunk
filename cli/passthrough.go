package main

import (
	"context"
	"diff-punk/cli/internal/git"
	"fmt"
	"os"
	"os/exec"
)

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
