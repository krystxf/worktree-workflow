# tmux

Creates a new tmux window (tab) for each worktree, named after the worktree folder.

## Setup

1. Copy `config.json` to `~/.config/worktree-workflow/config.json`
2. Copy `.worktree-workflow.json` to your project root
3. Run `wtw create <branch>` from inside a tmux session

## What it does

- `auto_open_editor` is `false` — no editor opens automatically
- The post-copy hook runs `tmux new-window` with:
  - `-n "$(basename "$PWD")"` — names the window after the worktree folder (e.g. `my-app--feature-x`)
  - `-c "$PWD"` — sets the working directory to the new worktree
- From there you can open your editor manually, run dev servers, etc.

## Example

For a repo `my-app` on branch `feature-x`, you get a tmux window named `my-app--feature-x` already `cd`'d into the worktree.
