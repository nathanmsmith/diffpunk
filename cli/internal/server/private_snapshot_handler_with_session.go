package server

import (
	"io/fs"
	"net"
	"net/http"
	"net/url"

	"diff-punk/cli/internal/git"
)

func snapshotHandlerWithSession(snapshot *git.Snapshot, assets fs.FS, address string, devServer *url.URL, session *viewerSession) http.Handler {
	var store *snapshotStore
	if session != nil && session.snapshots != nil {
		store = session.snapshots
	} else {
		store = newSnapshotStore(snapshot, false)
	}
	static := http.FileServer(http.FS(assets))
	if devServer != nil {
		static = devFrontend(devServer)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Restrict local hosts and reject cross-origin requests, including DNS rebinding.
		_, port, _ := net.SplitHostPort(address)
		if r.Host != address && r.Host != "localhost:"+port {
			http.NotFound(w, r)
			return
		}
		origin := r.Header.Get("Origin")
		site := r.Header.Get("Sec-Fetch-Site")
		if (origin != "" && origin != "http://"+r.Host) ||
			(site != "" && site != "same-origin" && site != "none") ||
			(r.Header.Get("Upgrade") != "" && origin != "http://"+r.Host) {
			http.Error(w, "Unexpected request origin", http.StatusForbidden)
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
			connectSources += " ws://" + r.Host
		}
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src "+scriptSources+"; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src "+connectSources+"; worker-src 'self' blob:; frame-ancestors 'none'; base-uri 'none'")
		if session != nil && r.URL.Path == "/session" {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			session.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/snapshot.json" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			if r.Method == http.MethodGet {
				_, _ = w.Write(store.read())
			}
			return
		}
		static.ServeHTTP(w, r)
	})
}
