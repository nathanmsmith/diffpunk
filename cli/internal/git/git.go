package git

import "diff-punk/cli/internal/git/private"

// File holds a file change and any available full-file context.
type File = private.File

// Snapshot is the frozen diff sent to the browser viewer.
type Snapshot = private.Snapshot
