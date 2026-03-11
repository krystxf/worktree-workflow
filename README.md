# Worktree Workflow

Stop stashing. Start branching like you mean it.

This script sets up [git worktrees](https://git-scm.com/docs/git-worktree) so you can work on multiple branches simultaneously — each in its own directory, each in its own editor window. No more `git stash` / `git stash pop`.

## What it does

1. Creates a worktree for the given branch next to your repo
2. Syncs gitignored files (`.env`, etc.) into the worktree via hard links
3. Runs any post-create commands you configure (e.g. `yarn install`)
4. Opens the worktree in your IDE

For a repo at `~/code/my-app`, running `worktree.sh develop` creates:

```
~/code/
  worktree.sh
  my-app/                        # your original repo
  my-app--worktrees/
    my-app--develop/             # the new worktree
```

- ✅ Switch branches instantly — no stashing, no losing context
- ✅ Each branch lives in its own folder with its own editor window
- ❌ Hard-linked `.env` files mean both worktrees share the same config
- ❌ Running dev servers for both at once requires manual port tweaking

## How to use

1. **Copy `worktree.sh`** somewhere close to your project
2. **Edit `worktree.sh`** — change the IDE (`cursor` → `code`, etc.), add post-create commands (`yarn install`, `pnpm install`, whatever your project needs), tweak `SYNC_EXCLUDES` if needed
3. **Run it:**

```bash
../worktree.sh <branch>
```

That's it. You're now working in a fresh worktree.
