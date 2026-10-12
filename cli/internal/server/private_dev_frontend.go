package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func devFrontend(target *url.URL) http.Handler {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			// The session index uses Vite's HTML transforms; its absolute asset
			// and WebSocket URLs are forwarded under devBase unchanged.
			if !strings.HasPrefix(r.In.URL.Path, devBase) {
				r.Out.URL.Path = devBase + strings.TrimPrefix(r.In.URL.Path, "/")
				r.Out.URL.RawPath = ""
			}
		},
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(w, "React dev server unavailable. Start npm run dev or restart diff-punk to reconnect.", http.StatusBadGateway)
		},
	}
}
