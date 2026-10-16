# diffpunk 🤖

<table>
  <tr>
    <td><img src="./diffpunk-browser.png" alt="diffpunk browser" /></td>
    <td><img src="./diffpunk-cli.png" alt="diffpunk CLI" /></td>
  </tr>
</table>

Diffpunk is a local, read-only browser diff viewer. Drop in replacement for `git diff`.

## Installation

With brew:

```sh
brew tap nathanmsmith/diffpunk https://github.com/nathanmsmith/diffpunk.git
brew install --HEAD diffpunk

# To update:
brew upgrade --fetch-HEAD nathanmsmith/diffpunk/diff-punk
```

Enable it as Git's diff pager:

```
git config --global pager.diff 'dp --pager'
git config --global diff.mnemonicPrefix true
git diff
git diff --staged
git diff main...HEAD -- src/
```

You can also run the standalone executable with the same arguments you would
pass to Git:

```sh
dp                    # unstaged changes
dp --staged           # staged changes
dp HEAD               # working tree compared with HEAD
dp main...HEAD        # branch changes since the merge base
dp -U10 -- src/       # more context, limited to a path
```

Standalone mode watches for changes and refreshes the diff automatically. Pager
mode displays the diff supplied by Git without watching.
