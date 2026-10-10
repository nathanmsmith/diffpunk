package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const devServerAddress = "http://127.0.0.1:5173"
const devBase = "/__diff_punk_dev/"
const devIdentity = "diff-punk-vite-v1"

// Release builds set this to false, including releases with the version "dev".
var development = "true"

// Only start tools from the checkout containing this development executable,
// never from the repository whose diff the user is viewing.
func devCheckout() string {
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return ""
	}
	root := filepath.Dir(executable)
	if filepath.Base(root) == "bin" {
		root = filepath.Dir(root)
	}
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return ""
	}
	var pkg struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(data, &pkg) != nil || pkg.Name != "diff-punk" {
		return ""
	}
	for _, name := range []string{"vite.config.mjs", "frontend/src/main.tsx"} {
		if info, err := os.Stat(filepath.Join(root, name)); err != nil || info.IsDir() {
			return ""
		}
	}
	return root
}

func prepareDevServer(ctx context.Context) (*url.URL, func(), error) {
	noop := func() {}
	if development != "true" {
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
		return nil, noop, fmt.Errorf("Vite is not installed; run npm ci in %s", root)
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
			return nil, noop, fmt.Errorf("Vite exited during startup: %v", processErr)
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

func detectDevServer(ctx context.Context, address string, enabled bool) *url.URL {
	if !enabled {
		return nil
	}
	target, err := url.Parse(address)
	if err != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address+devBase+"status", nil)
	if err != nil {
		return nil
	}
	client := &http.Client{
		Transport:     &http.Transport{Proxy: nil},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	defer client.CloseIdleConnections()
	res, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(res.Body, int64(len(devIdentity)+1)))
	if err != nil || res.StatusCode != http.StatusOK || string(body) != devIdentity {
		return nil
	}
	return target
}

func devFrontend(target *url.URL, prefix string) http.Handler {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			// The session index uses Vite's HTML transforms; its absolute asset
			// and WebSocket URLs are forwarded under devBase unchanged.
			if strings.HasPrefix(r.In.URL.Path, prefix) {
				r.Out.URL.Path = devBase + strings.TrimPrefix(r.In.URL.Path, prefix)
				r.Out.URL.RawPath = ""
			}
		},
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(w, "React dev server unavailable. Start npm run dev or restart diff-punk to reconnect.", http.StatusBadGateway)
		},
	}
}
