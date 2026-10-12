package server

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func prepareDevServer(ctx context.Context, development bool) (*url.URL, func(), error) {
	noop := func() {}
	if !development {
		return nil, noop, nil
	}
	if target := detectDevServer(ctx, devServerAddress, true); target != nil {
		return target, noop, nil
	}
	root := devCheckout()
	if root == "" {
		return nil, noop, nil
	}
	vite := filepath.Join(root, "node_modules", "vite", "bin", "vite.js")
	if _, err := os.Stat(vite); err != nil {
		return nil, noop, fmt.Errorf("vite is not installed; run npm ci in %s", root)
	}
	childCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(childCtx, "node", vite)
	cmd.Dir = root
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
	cmd.WaitDelay = 2 * time.Second
	fmt.Fprintln(os.Stderr, "diff punk: starting React dev server")
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, noop, fmt.Errorf("could not start Vite: %w", err)
	}
	exited := make(chan struct{})
	var processErr error
	go func() {
		processErr = cmd.Wait()
		close(exited)
	}()
	stop := func() {
		cancel()
		<-exited
	}
	startup, finish := context.WithTimeout(ctx, 10*time.Second)
	defer finish()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-exited:
			cancel()
			return nil, noop, fmt.Errorf("vite exited during startup: %v", processErr)
		case <-startup.Done():
			stop()
			return nil, noop, fmt.Errorf("waiting for Vite: %w", startup.Err())
		case <-ticker.C:
			if target := detectDevServer(startup, devServerAddress, true); target != nil {
				return target, stop, nil
			}
		}
	}
}
