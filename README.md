# diff punk

A local, read-only browser viewer for `git diff`. Enable it as Git's diff pager:

```sh
git config --global pager.diff 'diff-punk --pager'
git diff
git diff --staged
git diff main...HEAD -- src/
```

Git sends its output to diff punk, which opens the browser and keeps the terminal
occupied until you close the viewer or press **Ctrl-C**. This setup affects `git diff` only; other Git commands
keep their existing pager. Git retains control of comparison options and exit
status. `git --no-pager diff` bypasses the viewer, and redirected or piped Git
output keeps its normal behavior. To disable the integration:

```sh
git config --global --unset pager.diff
```

You can also run the standalone executable with the same arguments you would
pass to Git:

```sh
diff-punk                    # unstaged changes
diff-punk --staged           # staged changes
diff-punk HEAD               # working tree compared with HEAD
diff-punk main...HEAD        # branch changes since the merge base
diff-punk -U10 -- src/       # more context, limited to a path
```

The command opens your default browser and keeps a local server running on
`127.0.0.1:5555`. If port 5555 is already in use, startup fails. Press
**Ctrl-C** in the terminal to stop it, or close all viewer tabs. The server exits
five seconds after the last viewer disconnects, allowing time for a refresh to
reconnect. It waits for a viewer to connect before enabling automatic shutdown,
so you can also open the printed URL manually. Each invocation captures a snapshot;
subsequent edits do not change the viewer. An empty diff exits immediately.
Only one viewer can run at a time; close it or press Ctrl-C before starting another.

The viewer shows all files on one scrolling page, with a file navigation sidebar,
syntax highlighting, and light/dark colors matching your system. It defaults to
side-by-side diffs, with a unified toggle (unified by default on narrow screens).
**Show full file**, when available, expands a file's unchanged context;
**Show changes only** restores the requested Git context. Binary changes and
file metadata are shown as summaries.

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

### Homebrew

The formula in `Formula/diff-punk.rb` builds from `main` with Go and Node.js as
build dependencies. Until a tagged release is available, install with `--HEAD`.
For this private repository, first configure Git credentials with access to it
(for example, `gh auth setup-git` if you use GitHub CLI).

Once the formula is pushed to GitHub:

```sh
brew tap nathanmsmith/diffpunk https://github.com/nathanmsmith/diffpunk.git
brew install --HEAD nathanmsmith/diffpunk/diff-punk
diff-punk
```

To install the formula from this checkout before pushing it, run from the
repository root (create the local tap only if it does not already exist):

```sh
brew tap-new nathanmsmith/diffpunk
cp Formula/diff-punk.rb "$(brew --repository nathanmsmith/diffpunk)/Formula/diff-punk.rb"
brew install --HEAD nathanmsmith/diffpunk/diff-punk
```

The local formula still downloads the source from GitHub's `main` branch. To
update to the latest commit later:

```sh
brew upgrade --fetch-HEAD nathanmsmith/diffpunk/diff-punk
```

## Build from source

Requires Go 1.24+, Node.js 24+, npm, Git, and Make:

With [mise](https://mise.jdx.dev/), install the versions configured in
`mise.toml` (including golangci-lint for development checks), then build:

```sh
mise install
mise exec -- make build
./bin/diff-punk
```

If mise is activated in your shell, you can run `make build` directly.
With the required tools already on your `PATH`:

```sh
make build
./bin/diff-punk
```

Go sources live in `cli/`: `internal/git` captures and prepares diffs,
`internal/server` runs the browser viewer and Vite integration, and `assets/`
embeds the frontend. The Go module files remain at the repository root.

The frontend uses React, TypeScript, Vite, and `@pierre/diffs`. Go embeds the production
HTML, JavaScript, CSS, syntax grammars, themes, and third-party licenses. The viewer
loads everything locally and needs no internet connection at runtime.

## Git compatibility

With direct invocation, Git computes the patches; diff punk controls their
transport prefixes and disables terminal color for browser rendering. Normal
patch options, revisions, whitespace filters, diff algorithms, and pathspecs
retain Git's selection behavior.

In direct mode, reporting modes (`--stat`, `--name-only`, `--check`, and similar)
and specialized formats (`--raw`, `--word-diff`, `--binary`, combined diffs, external diff tools,
custom prefixes, and `--no-index`) stay in the terminal. `--quiet` and
`--exit-code` also pass through, preserving Git's output and exit status.
Unrecognized patch formats fall back to text output. `--help` invokes Git's diff
help; `--version` prints the diff punk version.

In pager mode, Git's supplied patch is the snapshot. Terminal color is removed
and mnemonic file prefixes are normalized to `a/` and `b/` for browser rendering.
Full-file context is reconstructed from the patch's Git
blob IDs and hunks before the server starts. Available blobs are read from Git;
missing sides are reconstructed in memory with exact hunk/context matching and
verified against the advertised blob IDs when available. The expanded view
preserves the supplied changed lines and adds verified unchanged context. It
does not substitute unverified working-tree contents or rerun the comparison.
Whitespace filtering, text conversions, missing blobs, or missing blob IDs can
make reconstruction unavailable; those files omit **Show full file** and still
display the supplied patch. Binary files are shown as summaries. Untracked files
are not included in either invocation mode; stage new files to include them in
`git diff --staged`.

Pager reporting output such as `--stat` and `--name-only`, and unsupported patch
formats, is printed unchanged without opening a browser. The pager also passes
stdin through unchanged when its stdout is redirected or piped. Configure
`pager.diff`, rather than `core.pager`: commit-grouped `git show` and `git log -p`
output is outside the initial pager scope.

Direct `diff-punk` invocation captures full-file context with a second Git
invocation before the server starts. Keep files stable during startup for the
patch and expanded context to represent the same state. After startup, both
views are frozen. Large files have no special limits in v1.

On Linux, opening the default browser uses `xdg-open`. If opening fails, the
terminal prints a URL you can open manually. Set `DIFF_PUNK_NO_OPEN=1` to run
without launching a browser (useful for tests or a headless environment).

## Development and checks

Build the CLI once with `make build`, then run it as usual:

```sh
./bin/diff-punk
./bin/diff-punk --staged
./bin/diff-punk main...HEAD
```

Development CLI builds automatically start this checkout's Vite server on
`127.0.0.1:5173` and open the viewer through Go on `127.0.0.1:5555`.
An already-running Vite server is reused. A server started by the CLI stops when
the CLI exits; to keep Vite running across CLI sessions, start `npm run dev` in
another terminal first.
React Fast Refresh preserves component state for compatible edits, and CSS updates
without reloading the page. Use the URL opened by the CLI; the Vite URL alone has
no diff snapshot. The CLI still captures a frozen diff at startup, and can be run
from any repository using the path to the executable.

The CLI locates frontend sources beside its executable (in the checkout root or
its `bin` directory), so it also works when launched from another repository.
Outside the checkout, development binaries attach to an existing Vite server or
use embedded assets. If Node.js or installed Vite dependencies are unavailable,
startup explains the problem and falls back to embedded assets. An unrelated
service on port 5173 also causes a fallback. If Vite stops during a session,
restart the CLI to start it again. Release archives always use embedded assets.
Go edits require rebuilding the CLI with `go build -o bin/diff-punk ./cli` and
restarting it. Run `make build` to rebuild both the frontend and the CLI.

`npm run build` uses Vite to generate production assets and third-party license
notices in `cli/assets/dist` before Go embeds them.

```sh
make test                 # build, lint, format, typecheck, Go race, DOM and CLI checks
make browser-test         # install Chromium and check production + Fast Refresh
make lint-go              # build embedded assets and check Go code and formatting
npm run lint:fix          # fix automatically correctable ESLint issues
npm run format            # format JS/TS, CSS, HTML, JSON, Markdown, and YAML
npm run test:dev           # check dev integration (requires built CLI and browser)
```

ESLint checks the frontend TypeScript/TSX and Node.js scripts with recommended rules.
Prettier handles formatting; `npm run lint` and `npm run format:check` check without
changing files. Both checks also run in CI through `make test`.

Go checks use golangci-lint v2.13.2 with `errcheck`, `govet`, `ineffassign`,
`staticcheck`, `unused`, `gofmt`, and `goimports`. `mise install` provides the same
version used in CI. Alternatively, install it with the
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
both color schemes, and that all assets load locally. Dev integration checks edit
an isolated frontend copy to verify React/CSS refresh, state preservation, frozen
snapshots, automatic Vite startup and shutdown, release isolation, and fallback
to embedded assets.

```sh
make release              # build four standalone release archives and checksums
./scripts/release.sh v0.1.0  # after make frontend, stamp a specific version
```

Pushing a `v*` tag runs the release checks and publishes the archives to GitHub
Releases. Third-party license notices are embedded in the viewer and included in
each archive.
