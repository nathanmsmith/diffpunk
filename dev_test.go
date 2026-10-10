package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"
)

func TestDetectDevServer(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		enabled    bool
		want       bool
	}{
		{"ours", devIdentity, 200, true, true},
		{"unrelated", "another Vite app", 200, true, false},
		{"extra data", devIdentity + "extra", 200, true, false},
		{"error", devIdentity, 500, true, false},
		{"redirect", devIdentity, 302, true, false},
		{"release", devIdentity, 200, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !tc.enabled {
					t.Error("release build probed the dev server")
				}
				if r.URL.Path != devBase+"status" {
					t.Errorf("unexpected probe path: %s", r.URL.Path)
				}
				w.Header().Set("Location", "/redirected")
				w.WriteHeader(tc.status)
				_, _ = fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			got := detectDevServer(context.Background(), server.URL, tc.enabled)
			if (got != nil) != tc.want {
				t.Fatalf("detected %v, want %v", got, tc.want)
			}
		})
	}
	t.Run("unavailable", func(t *testing.T) {
		server := httptest.NewServer(http.NotFoundHandler())
		server.Close()
		if detectDevServer(context.Background(), server.URL, true) != nil {
			t.Fatal("detected a stopped server")
		}
	})
	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		defer server.Close()
		if detectDevServer(context.Background(), server.URL, true) != nil {
			t.Fatal("detected an unresponsive server")
		}
	})
}

func TestDevFrontendProxy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "127.0.0.1:1234" {
			t.Error("proxy did not rewrite the Host")
		}
		_, _ = fmt.Fprint(w, r.URL.RequestURI())
	}))
	defer upstream.Close()
	target, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := &Snapshot{Files: []File{{Name: "foo.go", Patch: "frozen"}}}
	handler := snapshotHandler(snapshot, fstest.MapFS{}, "127.0.0.1:1234", "/secret/", target)
	snapshot.Files[0].Patch = "changed"
	for _, tc := range []struct {
		path, host, method, origin, upgrade string
		status                              int
		body                                string
	}{
		{"/secret/", "127.0.0.1:1234", "GET", "", "", 200, devBase},
		{devBase + "src/App.tsx?t=42", "127.0.0.1:1234", "GET", "", "", 200, devBase + "src/App.tsx?t=42"},
		{"/secret/snapshot.json", "127.0.0.1:1234", "GET", "", "", 200, `"patch":"frozen"`},
		{"/secret/snapshot.json", "127.0.0.1:1234", "HEAD", "", "", 200, ""},
		{"/snapshot.json", "127.0.0.1:1234", "GET", "", "", 404, ""},
		{"/wrong/snapshot.json", "127.0.0.1:1234", "GET", "", "", 404, ""},
		{"/secret/snapshot.json", "evil.example", "GET", "", "", 404, ""},
		{devBase + "src/App.tsx", "evil.example", "GET", "", "", 404, ""},
		{devBase, "127.0.0.1:1234", "POST", "", "", 405, ""},
		{devBase, "127.0.0.1:1234", "GET", "http://evil.example", "websocket", 403, ""},
	} {
		t.Run(tc.method+tc.path+tc.host+tc.origin, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "http://127.0.0.1:1234"+tc.path, nil)
			req.Host = tc.host
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Upgrade", tc.upgrade)
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != tc.status || !strings.Contains(res.Body.String(), tc.body) {
				t.Fatalf("got %d %q; want %d containing %q", res.Code, res.Body.String(), tc.status, tc.body)
			}
			if tc.method == "HEAD" && res.Body.Len() != 0 {
				t.Fatal("HEAD response has a body")
			}
		})
	}
	upstream.Close()
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest("GET", "http://127.0.0.1:1234/secret/", nil))
	if res.Code != http.StatusBadGateway || !strings.Contains(res.Body.String(), "npm run dev") {
		t.Fatalf("missing dev server recovery message: %d %s", res.Code, res.Body.String())
	}
}
