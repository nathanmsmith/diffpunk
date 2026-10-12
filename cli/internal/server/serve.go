package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"diff-punk/cli/assets"
	"diff-punk/cli/internal/git"
)

// Serve runs the local viewer until its session closes or the context is canceled.
func Serve(ctx context.Context, snapshot *git.Snapshot, development bool, refreshes ...Refresh) error {
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
	address := listener.Addr().String()
	url := "http://localhost:5555/"
	var refresh Refresh
	if len(refreshes) > 0 {
		refresh = refreshes[0]
	}
	store := newSnapshotStore(snapshot, refresh != nil)
	session := &viewerSession{ctx: ctx, stop: cancelSession, snapshots: store}
	if refresh != nil {
		go store.watch(ctx, refresh)
	}
	server := &http.Server{
		Handler:           snapshotHandlerWithSession(snapshot, assetFS, address, devServer, session),
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
