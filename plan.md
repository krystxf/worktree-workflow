# wtw — Worktree Workflow

Feature roadmap for the `wtw` CLI tool.

## Implemented

- [x] `wtw create <branch>` — create worktree with TUI spinner + log streaming
- [x] `wtw ps` — interactive worktree list, select to open in editor
- [x] `wtw rm <branch>` — remove a worktree
- [x] Global config (`~/.config/worktree-workflow/config.json`) — editor, auto-open, naming
- [x] Per-project config (`.worktree-workflow.json`) — sync_ignored, excludes, post-copy hooks
- [x] Sync gitignored files via hard links (rsync)
- [x] Post-copy hooks (e.g. `yarn install`)
- [x] golangci-lint + gofumpt setup

## Planned

- [ ] `wtw init` — generate `.worktree-workflow.json` interactively
- [ ] Worktree cleanup of stale/orphaned worktrees
- [ ] Tab completion for branch names
- [ ] Config validation with helpful error messages
