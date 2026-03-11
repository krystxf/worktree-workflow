# wtw — Go CLI

## Install

### Homebrew

```bash
brew tap krystxf/wtw
brew install wtw
```

## Configuration

### Global config

Create `~/.config/worktree-workflow/config.json`:

```json
{
  "editor": "cursor",
  "auto_open_editor": true,
  "naming": {
    "worktree_dir_suffix": "--worktrees",
    "branch_separator": "--"
  }
}
```

| Field                        | Default         | Description                                               |
| ---------------------------- | --------------- | --------------------------------------------------------- |
| `editor`                     | `"cursor"`      | Command to open the worktree (e.g. `code`, `nvim`, `zed`) |
| `auto_open_editor`           | `true`          | Open editor automatically after creating a worktree       |
| `naming.worktree_dir_suffix` | `"--worktrees"` | Suffix for the parent worktree directory                  |
| `naming.branch_separator`    | `"--"`          | Separator between repo name and branch name               |

All fields are optional — missing values use the defaults above.

### Per-project config

Create `.worktree-workflow.json` in your project root:

```json
{
  "sync_ignored": true,
  "sync_excludes": ["node_modules"],
  "post_copy_hooks": ["npm install"]
}
```

| Field             | Default | Description                                              |
| ----------------- | ------- | -------------------------------------------------------- |
| `sync_ignored`    | `true`  | Sync gitignored files (`.env`, etc.) via hard links      |
| `sync_excludes`   | `[]`    | Patterns to exclude from sync                            |
| `post_copy_hooks` | `[]`    | Shell commands to run in the new worktree after creation |

## Examples

See the [`examples/`](examples/) directory:

- [`cursor-npm`](examples/cursor-npm/) — Cursor editor + `npm install`
- [`tmux`](examples/tmux/) — tmux window per worktree

## Development

### Prerequisites

- [Go](https://go.dev/dl/) 1.21+
- git
- rsync (pre-installed on macOS/Linux)

### Build

```bash
git clone https://github.com/krystof/worktree-workflow.git
cd worktree-workflow
go build -o wtw .
```

This produces a `wtw` binary in the current directory.

### Run locally

```bash
# Create a worktree
./wtw create feature-branch

# List worktrees (interactive picker)
./wtw ls

# Remove a worktree
./wtw rm feature-branch
```

### Make targets

```bash
# Build
make build

# Format code
make fmt

# Lint (requires golangci-lint)
make lint

# Clean binary
make clean
```

### Install golangci-lint

```bash
brew install golangci-lint
```

Or see [golangci-lint.run](https://golangci-lint.run/welcome/install/) for other methods.
