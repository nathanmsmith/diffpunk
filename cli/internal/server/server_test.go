package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"diff-punk/cli/internal/git"
)

func TestServerServesFrozenSnapshotAndAssets(t *testing.T) {
	snapshot := &git.Snapshot{Files: []git.File{{Name: "foo.go", Patch: "snapshot"}}}
	assets := fstest.MapFS{"index.html": {Data: []byte("viewer")}, "assets/main.js": {Data: []byte("javascript")}}
	handler := snapshotHandler(snapshot, assets, "127.0.0.1:1234", nil)
	// The handler serializes once; later changes cannot alter the snapshot.
	snapshot.Files[0].Patch = "changed"
	for _, tc := range []struct {
		url, host, body string
		status          int
	}{
		{"http://127.0.0.1:1234/", "127.0.0.1:1234", "viewer", 200},
		{"http://127.0.0.1:1234/assets/main.js", "127.0.0.1:1234", "javascript", 200},
		{"http://127.0.0.1:1234/snapshot.json", "evil.example", "", 404},
		{"http://localhost:1234/", "localhost:1234", "viewer", 200},
	} {
		req := httptest.NewRequest("GET", tc.url, nil)
		req.Host = tc.host
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != tc.status {
			t.Errorf("%s: status %d", tc.url, res.Code)
		}
		if tc.body != "" && res.Body.String() != tc.body {
			t.Errorf("%s: body %q", tc.url, res.Body.String())
		}
	}
	req := httptest.NewRequest("GET", "http://127.0.0.1:1234/snapshot.json", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	var got git.Snapshot
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Files[0].Patch != "snapshot" {
		t.Fatal("snapshot was mutated")
	}
	if res.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("snapshot permits cross-origin reads")
	}
}
