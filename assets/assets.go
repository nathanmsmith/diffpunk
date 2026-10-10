package assets

import "embed"

// Assets includes the production JavaScript, syntax grammars, themes, and CSS.
// Build the frontend with npm run build before compiling the Go executable.
//
//go:embed all:dist
var Assets embed.FS
