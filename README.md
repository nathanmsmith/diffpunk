# diff punk

A local, read-only browser viewer for `git diff`. Run the standalone executable
with the same arguments you would pass to Git:

```sh
diff-punk                    # unstaged changes
diff-punk --untracked        # initially show untracked files
diff-punk --staged           # staged changes
diff-punk HEAD               # working tree compared with HEAD
diff-punk main...HEAD        # branch changes since the merge base
diff-punk -U10 -- src/       # more context, limited to a path
```

The command opens your default browser and keeps a local server running. Press
**Ctrl-C** in the terminal to stop it. Each invocation captures a snapshot;
subsequent edits do not change the viewer. An empty diff exits immediately unless eligible untracked files are present.

The viewer shows all files on one scrolling page, with a file navigation sidebar,
syntax highlighting, and light/dark colors matching your system. It defaults to
side-by-side diffs, with a unified toggle (unified by default on narrow screens).
**Show full file** expands a file's unchanged context; **Show changes only** restores
the requested Git context. Binary changes and file metadata are shown as summaries.

Files are ordered by directory depth, then path. Changed tests follow their
matching implementation immediately, including tests in a separate directory.
Built-in conventions recognize `.test`, `.spec`, `_test`, `_spec`, `test_`,
`Test`, `Tests`, and `Spec` filenames, plus `test/`, `tests/`, `spec/`, `specs/`,
and `__tests__/` directories. Source/test paths are matched within modules;
ambiguous tests remain unpaired. Unpaired tests come last, sorted by depth and path.

## Installation

Release builds are standalone executables for macOS and Linux, on ARM64 and x86-64.
Download the appropriate archive from this repository's GitHub Releases, verify
its checksum, extract it, and put `diff-punk` on your `PATH`. Git must be installed;
Go and Node.js are only needed to build from source.

## Build from source

Requires Go 1.24+, Node.js 24+, npm, Git, and Make:

```sh
make build
./bin/diff-punk
```

The frontend uses React, TypeScript, and `@pierre/diffs`. Go embeds the production
HTML, JavaScript, CSS, syntax grammars, themes, and third-party licenses. The viewer
loads everything locally and needs no internet connection at runtime.

## Git compatibility

Git computes the patches; diff punk controls their transport prefixes and disables
terminal color for browser rendering. Normal patch options, revisions, whitespace
filters, diff algorithms, and pathspecs retain Git's selection behavior.

Reporting modes (`--stat`, `--name-only`, `--check`, and similar) and specialized
formats (`--raw`, `--word-diff`, `--binary`, combined diffs, external diff tools,
custom prefixes, and `--no-index`) stay in the terminal. `--quiet` and
`--exit-code` also pass through, preserving Git's output and exit status.
Unrecognized patch formats fall back to text output. `--help` invokes Git's diff
help; `--version` prints the diff punk version.

Full-file context is captured with a second Git invocation before the server
starts. Keep files stable during startup for the patch and expanded context to
represent the same state. After startup, both views are frozen. Large files have
no special limits in v1. Working-tree views capture untracked files too, respecting
Git ignore rules and supplied pathspecs. **Show untracked files** toggles their
visibility; they appear as complete added-file diffs labeled **Untracked**.
They are hidden by default; `--untracked` shows them initially. The browser setting
lasts only for that page session. If only untracked files exist, the viewer opens
with a prompt to enable the toggle. Staged and commit-to-commit comparisons omit
untracked files and hide the toggle. Reporting modes retain Git's normal behavior;
`--untracked` is stripped before passing those commands to Git.

On Linux, opening the default browser uses `xdg-open`. If opening fails, the
terminal prints a URL you can open manually. Set `DIFF_PUNK_NO_OPEN=1` to run
without launching a browser (useful for tests or a headless environment).

## Development and checks

```sh
make test                 # build, lint, format, typecheck, Go race, DOM and CLI checks
make browser-test         # install Chromium and run browser integration checks
make lint-go              # build embedded assets and check Go code and formatting
npm run lint:fix          # fix automatically correctable ESLint issues
npm run format            # format JS/TS, CSS, HTML, JSON, Markdown, and YAML
```

ESLint checks the frontend TypeScript/TSX and Node.js scripts with recommended rules.
Prettier handles formatting; `npm run lint` and `npm run format:check` check without
changing files. Both checks also run in CI through `make test`.

Go checks use golangci-lint v2.13.2 with `errcheck`, `govet`, `ineffassign`,
`staticcheck`, `unused`, `gofmt`, and `goimports`. Install the same version used in CI with the
[official binary installer](https://golangci-lint.run/docs/welcome/install/local/):

```sh
curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v2.13.2
```

Add `$(go env GOPATH)/bin` to your `PATH`, then run `make lint-go` or `make test`.
Use `golangci-lint fmt` to format Go files and organize imports. Go checks run on both CI platforms and
before releases through `make test`.

The checks use temporary repositories and cover staging, revision comparisons,
path filters, renames, deletion, binary files, deterministic pairing, frozen
full-file context, terminal output/exit codes, browser launching, and shutdown.
CI runs the checks on macOS and Linux. Browser tests also check responsive layouts,
both color schemes, and that all assets load locally.

```sh
make release              # build four standalone release archives and checksums
./scripts/release.sh v0.1.0  # after make frontend, stamp a specific version
```

Pushing a `v*` tag runs the release checks and publishes the archives to GitHub
Releases. Third-party license notices are embedded in the viewer and included in
each archive.
