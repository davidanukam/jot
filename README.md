# Jot

A commit-message journal CLI. Jot short notes as you work; assemble them into a
properly formatted commit message right before `git commit`.

## Install

```sh
go build -o jot .
```

Put the binary on your `PATH` (e.g. copy it to `~/bin` or `/usr/local/bin`).

Cross-compile without network access:

```sh
GOOS=linux GOARCH=amd64 go build -o jot .
GOOS=darwin GOARCH=arm64 go build -o jot .
GOOS=windows GOARCH=amd64 go build -o jot.exe .
```

## Refined flow

```sh
# Start working in a repo
cd my-project

# Set up the post-commit hook (optional; first jot write installs it too)
jot init

# Log notes as you go
jot write "Added a new function to render circles"
jot write "typo note"
jot undo
jot write "Refactored main pipeline and packages"

# Review what's pending for the next commit
jot read

# Tweak notes before committing
jot edit 0 "Added circle rendering helper"
jot main 1
jot rm 0

# Stage, copy, and commit
jot update
jot copy
jot paste

# Push to remote (e.g. first push to a new branch)
jot push -u origin main

# After commit, pending notes are cleared automatically
jot read   # → No messages have been made for the next commit

# Browse past commits
jot log

# Drop pending notes without committing
jot clear

jot help
jot version
```

## Commands

| Command | Description |
|---------|-------------|
| `jot write "message"` | Log a note (multi-line quoted text supported) |
| `jot write --no-hyphen "message"` | Log multi-line without auto-hyphens on sub-lines |
| `jot read` | List pending notes for the next commit (git-log-style dates) |
| `jot log` | Show past commit messages (`git log` passthrough) |
| `jot main <n>` | Set subject line (index from `read`) |
| `jot copy` / `jot -c` | Format and copy to clipboard |
| `jot copy --preview` | Print formatted message only |
| `jot paste` | Copy to clipboard and `git commit` |
| `jot paste -p` | Preview message, confirm `[y/N]`, then copy and commit |
| `jot update` | Stage all changes (`git add .`) |
| `jot push` | Push to remote (`git push` passthrough) |
| `jot rm <n>` | Remove a note |
| `jot edit <n> <text>` | Edit a note in place |
| `jot undo` | Remove the last note |
| `jot clear` | Wipe all notes (`-y` to skip prompt) |
| `jot init` | Install/repair the post-commit hook |

## Storage

Notes live at `<git-dir>/jot/log.json` (inside `.git/`, never tracked). Each
git worktree gets its own log. Legacy data at `.git/gut/log.json` is read
automatically if present.

## Clipboard

- **macOS**: `pbcopy`
- **Windows**: `clip`
- **Linux**: `wl-copy`, `xclip`, or `xsel` (first found on `PATH`)

If clipboard access fails, the formatted message is still printed to stdout.
