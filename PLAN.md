The plan: Build v1 as a Go executable with an embedded plain-TypeScript viewer using @pierre/diffs.

The agreed behavior:

• Preserve git diff semantics, including unstaged changes by default.
• Open ordinary patches in the default browser immediately.
• Run a localhost server in the foreground until Ctrl-C.
• Show a read-only snapshot in one scrolling page.
• Default to side-by-side, with a unified toggle.
• Allow expanding unchanged context to the whole file using captured file versions.
• Pass reporting options and specialized formats through to Git in the terminal.
• Distribute standalone macOS and Linux binaries through GitHub Releases.

File order is deterministic: sort implementation files by directory depth, then path; place
matching changed tests immediately after each implementation using built-in conventions. Put
unpaired tests last, sorted by depth and path.

I recommend building it in three steps:

1. Complete the basic flow: invoke Git, capture a patch, serve the embedded viewer, launch the
   browser, and shut down cleanly.

2. Add review behavior: deterministic pairing/order, layout toggle, and full-file expansion.
3. Verify and package: exercise unstaged, staged, revision, and path-filtered diffs; cover
   renames, deletions, binary files, and terminal passthrough; produce release binaries.

Keep v1 small: defer AI ranking, editing, live updates, custom pairing configuration,
Homebrew, and special large-file handling.
