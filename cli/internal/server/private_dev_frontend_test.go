package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"

	"diff-punk/cli/internal/git"
)

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
	snapshot := &git.Snapshot{Files: []git.File{{Name: "foo.go", Patch: "frozen"}}}
	handler := snapshotHandler(snapshot, fstest.MapFS{}, "127.0.0.1:1234", target)
	snapshot.Files[0].Patch = "changed"
	for _, tc := range []struct {
		path, host, method, origin, upgrade string
		status                              int
		body                                string
	}{
		{"/", "127.0.0.1:1234", "GET", "", "", 200, devBase},
		{devBase + "src/App.tsx?t=42", "127.0.0.1:1234", "GET", "", "", 200, devBase + "src/App.tsx?t=42"},
		{"/snapshot.json", "127.0.0.1:1234", "GET", "", "", 200, `"patch":"frozen"`},
		{"/snapshot.json", "127.0.0.1:1234", "HEAD", "", "", 200, ""},
		{"/snapshot.json", "evil.example", "GET", "", "", 404, ""},
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
	handler.ServeHTTP(res, httptest.NewRequest("GET", "http://127.0.0.1:1234/", nil))
	if res.Code != http.StatusBadGateway || !strings.Contains(res.Body.String(), "npm run dev") {
		t.Fatalf("missing dev server recovery message: %d %s", res.Code, res.Body.String())
	}
}
