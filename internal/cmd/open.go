package cmd

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/krystof/worktree-workflow/internal/config"
	"github.com/krystof/worktree-workflow/internal/git"
	"github.com/krystof/worktree-workflow/internal/ui"
)

var openCmd = &cobra.Command{
	Use:   "open [branch]",
	Short: "Open a worktree in the configured editor",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := git.MainRoot()
		if err != nil {
			return fmt.Errorf("not a git repository")
		}

		globalCfg, err := config.LoadGlobal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load global config: %s\n", err)
		}

		if len(args) == 1 {
			branch := args[0]
			worktreeDir := git.WorktreeDir(root, globalCfg.Naming.WorktreeDirSuffix, globalCfg.Naming.BranchSeparator, branch)
			return openEditor(globalCfg, worktreeDir)
		}

		// No branch given — interactive picker
		worktrees, err := git.WorktreeList()
		if err != nil {
			return err
		}

		if len(worktrees) == 0 {
			fmt.Println("No worktrees found.")
			return nil
		}

		model := ui.NewPickerModel("Open worktree", worktrees, ui.OpenEditorAction(globalCfg), root, "", "Opening")
		p := tea.NewProgram(model, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		if m, ok := finalModel.(ui.PickerModel); ok {
			if path := m.SelectedPath(); path != "" {
				return openEditor(globalCfg, path)
			}
		}

		return nil
	},
}

func openEditor(globalCfg config.GlobalConfig, worktreeDir string) error {
	bin, editorArgs := globalCfg.EditorArgs(worktreeDir)
	c := exec.Command(bin, editorArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}
	return nil
}
