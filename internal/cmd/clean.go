package cmd

import (
	"errors"
	"fmt"

	"github.com/krystof/worktree-workflow/internal/git"
	"github.com/spf13/cobra"
)

func cleanWorktrees(cmd *cobra.Command) error {
	current, err := git.Root()
	if err != nil {
		return err
	}
	worktrees, err := git.WorktreeList()
	if err != nil {
		return err
	}
	removed := 0
	var failures []error
	for i, w := range worktrees {
		reason := ""
		switch {
		case i == 0:
			reason = "main worktree"
		case w.Path == current:
			reason = "current worktree"
		default:
			reason, err = git.CleanReason(w)
			if err != nil {
				failures = append(failures, fmt.Errorf("%s: %w", w.Path, err))
				cmd.PrintErrf("Skipped %s: %s\n", w.Path, err)
				continue
			}
		}
		if reason != "" {
			cmd.Printf("Skipped %s: %s\n", w.Path, reason)
			continue
		}
		// Never force removal, even when the global --force flag is set.
		if _, err := git.WorktreeRemove(w.Path, false); err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", w.Path, err))
			continue
		}
		removed++
		cmd.Printf("Removed %s\n", w.Path)
	}
	cmd.Printf("Removed %d synced worktree(s).\n", removed)
	return errors.Join(failures...)
}
