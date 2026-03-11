# Cursor + npm

Opens new worktrees in Cursor and runs `npm install` automatically.

## Setup

1. Copy `config.json` to `~/.config/worktree-workflow/config.json`
2. Copy `.worktree-workflow.json` to your project root

## What it does

- Opens Cursor on the new worktree directory
- Syncs gitignored files (`.env`, etc.) via hard links, skipping `node_modules` and `.next`
- Runs `npm install` in the new worktree so dependencies are ready
