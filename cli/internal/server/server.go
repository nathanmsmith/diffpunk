package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"diff-punk/cli/assets"
	"diff-punk/cli/internal/git"
)

// Serve runs the local viewer until its session closes or the context is canceled.
func Serve(ctx context.Context, snapshot *git.Snapshot, development bool) error {
	ctx, cancelSession := context.WithCancel(ctx)
	defer cancelSession()
	listener, err := net.Listen("tcp", "127.0.0.1:5555")
	if err != nil {
		if errors.Is(err, syscall.EADDRINUSE) {
			return fmt.Errorf("viewer port 5555 is already in use; close the active viewer or press Ctrl-C before starting another")
		}
		return err
	}
	// Serve closes the listener; also clean up if setup fails before Serve starts.
	defer func() { _ = listener.Close() }()
	devServer, stopDevServer, err := prepareDevServer(ctx, development)
	defer stopDevServer()
	if ctx.Err() != nil {
		return nil
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "diff punk:", err, "(using embedded assets)")
	}
	assetFS, err := fs.Sub(assets.Assets, "dist")
	if err != nil {
		return err
	}
	if devServer == nil {
		if _, err := fs.Stat(assetFS, "index.html"); err != nil {
			return fmt.Errorf("viewer assets missing; run npm run build before go build")
		}
	}
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return err
	}
	prefix := "/" + hex.EncodeToString(token[:]) + "/"
	address := listener.Addr().String()
	url := "http://" + address + prefix
	session := &viewerSession{ctx: ctx, stop: cancelSession}
	server := &http.Server{
		Handler:           snapshotHandlerWithSession(snapshot, assetFS, address, prefix, devServer, session),
		ReadHeaderTimeout: 5 * time.Second,
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	if devServer != nil {
		fmt.Fprintln(os.Stderr, "diff punk: React dev server detected (Fast Refresh enabled)")
	}
	fmt.Fprintf(os.Stderr, "diff punk: %s\nClose the viewer to stop, or press Ctrl-C.\n", url)
	if os.Getenv("DIFF_PUNK_NO_OPEN") != "1" {
		if err := openBrowser(url); err != nil {
			fmt.Fprintln(os.Stderr, "diff punk: could not open browser; open the URL above:", err)
		}
	}
	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func snapshotHandler(snapshot *git.Snapshot, assets fs.FS, address, prefix string, devServer *url.URL) http.Handler {
	return snapshotHandlerWithSession(snapshot, assets, address, prefix, devServer, nil)
}

func snapshotHandlerWithSession(snapshot *git.Snapshot, assets fs.FS, address, prefix string, devServer *url.URL, session *viewerSession) http.Handler {
	data, _ := json.Marshal(snapshot)
	static := http.StripPrefix(prefix, http.FileServer(http.FS(assets)))
	if devServer != nil {
		static = devFrontend(devServer, prefix)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A loopback listener, exact Host, and unguessable URL keep repository
		// contents inaccessible to unrelated websites and DNS rebinding requests.
		devRequest := devServer != nil && strings.HasPrefix(r.URL.Path, devBase)
		if r.Host != address || (!strings.HasPrefix(r.URL.Path, prefix) && !devRequest) {
			http.NotFound(w, r)
			return
		}
		if devServer != nil && r.Header.Get("Upgrade") != "" && r.Header.Get("Origin") != "http://"+address {
			http.Error(w, "Unexpected WebSocket origin", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		scriptSources := "'self' 'wasm-unsafe-eval'"
		connectSources := "'self'"
		if devServer != nil {
			// Vite injects an inline React Refresh preamble and uses WebSockets.
			scriptSources += " 'unsafe-inline'"
			connectSources += " ws://" + address
		}
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src "+scriptSources+"; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src "+connectSources+"; worker-src 'self' blob:; frame-ancestors 'none'; base-uri 'none'")
		if session != nil && r.URL.Path == prefix+"session" {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			session.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == prefix+"snapshot.json" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			if r.Method == http.MethodGet {
				_, _ = w.Write(data)
			}
			return
		}
		static.ServeHTTP(w, r)
	})
}

func openBrowser(url string) error {
	program := "xdg-open"
	if runtime.GOOS == "darwin" {
		program = "open"
	}
	cmd := exec.Command(program, url)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Fprintln(os.Stderr, "diff punk: browser launcher failed; open the printed URL:", err)
		}
	}()
	return nil
}
