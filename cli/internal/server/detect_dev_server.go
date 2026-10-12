package server

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

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
