package server

import (
	"diff-punk/cli/internal/git"
	"io/fs"
	"net/http"
	"net/url"
)

func snapshotHandler(snapshot *git.Snapshot, assets fs.FS, address string, devServer *url.URL) http.Handler {
	return snapshotHandlerWithSession(snapshot, assets, address, devServer, nil)
}
