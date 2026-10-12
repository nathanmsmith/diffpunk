package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
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
