package ui

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/krystof/worktree-workflow/internal/config"
	gitpkg "github.com/krystof/worktree-workflow/internal/git"
)

// OpenEditorAction returns a PickerAction that signals the picker to quit so the
// editor can be launched after the TUI exits (avoids terminal state conflicts).
func OpenEditorAction(globalCfg config.GlobalConfig) PickerAction {
	return func(item worktreeItem) tea.Cmd {
		return func() tea.Msg {
			return actionDoneMsg{result: item.path}
		}
	}
}

// RemoveWorktreeAction returns a PickerAction that removes the selected worktree and prunes.
// If remove fails due to modified/untracked files, it sends needsForceMsg to prompt the user.
func RemoveWorktreeAction(force bool) PickerAction {
	return func(item worktreeItem) tea.Cmd {
		return func() tea.Msg {
			_, err := gitpkg.WorktreeRemove(item.path, force)
			if err != nil {
				if !force && errors.Is(err, gitpkg.ErrWorktreeModified) {
					return needsForceMsg{item: item}
				}
				return actionDoneMsg{err: err}
			}

			_, pruneErr := gitpkg.WorktreePrune()
			if pruneErr != nil {
				return actionDoneMsg{result: fmt.Sprintf("Removed worktree '%s' (prune warning: %s)", item.branch, pruneErr)}
			}

			return actionDoneMsg{result: fmt.Sprintf("Removed worktree '%s'", item.branch)}
		}
	}
}

func forceRemoveAction(item worktreeItem) tea.Cmd {
	return func() tea.Msg {
		_, err := gitpkg.WorktreeRemove(item.path, true)
		if err != nil {
			return actionDoneMsg{err: err}
		}

		_, pruneErr := gitpkg.WorktreePrune()
		if pruneErr != nil {
			return actionDoneMsg{result: fmt.Sprintf("Force removed worktree '%s' (prune warning: %s)", item.branch, pruneErr)}
		}

		return actionDoneMsg{result: fmt.Sprintf("Force removed worktree '%s'", item.branch)}
	}
}
