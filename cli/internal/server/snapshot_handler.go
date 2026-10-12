package server

import (
	"io/fs"
	"net/http"
	"net/url"

	"diff-punk/cli/internal/git"
)

func snapshotHandler(snapshot *git.Snapshot, assets fs.FS, address string, devServer *url.URL) http.Handler {
	return snapshotHandlerWithSession(snapshot, assets, address, devServer, nil)
}
