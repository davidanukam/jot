# Gut

A commit-message journal CLI. Jot short notes as you work; assemble them into a
properly formatted commit message right before `git commit`.

## Install

```sh
go build -o gut .
```

Put the binary on your `PATH` (e.g. copy it to `~/bin` or `/usr/local/bin`).

Cross-compile without network access:

```sh
GOOS=linux GOARCH=amd64 go build -o gut .
GOOS=darwin GOARCH=arm64 go build -o gut .
GOOS=windows GOARCH=amd64 go build -o gut.exe .
```

## Refined flow

```sh
# Start working in a repo
cd my-project

# Log notes as you go (hook installs automatically on first use)
gut "Added a new function to render circles"
gut "Refactored main pipeline and packages"

# Review what's pending
gut show

# Pick which note becomes the commit subject
gut main 1

# Copy formatted message to clipboard (also prints to stdout)
gut copy

# Commit using the formatted message
git commit -m "$(gut copy --preview)"

# After commit, notes are cleared automatically by the post-commit hook
gut show   # → no pending gut notes since the last commit
```

## Commands

| Command | Description |
|---------|-------------|
| `gut "<message>"` | Log a note |
| `gut show` | List pending notes with timestamps |
| `gut main <n>` | Set subject line (index from `show`) |
| `gut copy` / `gut -c` | Format and copy to clipboard |
| `gut copy --preview` | Print formatted message only |
| `gut rm <n>` | Remove a note |
| `gut edit <n> <text>` | Edit a note in place |
| `gut undo` | Remove the last note |
| `gut clear` | Wipe all notes (`-y` to skip prompt) |
| `gut init` | Install/repair the post-commit hook |

## Storage

Notes live at `<git-dir>/gut/log.json` (inside `.git/`, never tracked). Each
git worktree gets its own log.

## Clipboard

- **macOS**: `pbcopy`
- **Windows**: `clip`
- **Linux**: `wl-copy`, `xclip`, or `xsel` (first found on `PATH`)

If clipboard access fails, the formatted message is still printed to stdout.
