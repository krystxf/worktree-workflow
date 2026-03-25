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

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List worktrees and select one to open",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := git.MainRoot()
		if err != nil {
			return fmt.Errorf("not a git repository")
		}

		globalCfg, err := config.LoadGlobal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load global config: %s\n", err)
		}

		worktrees, err := git.WorktreeList()
		if err != nil {
			return err
		}

		if len(worktrees) == 0 {
			fmt.Println("No worktrees found.")
			return nil
		}

		model := ui.NewPickerModel("Worktrees", worktrees, ui.OpenEditorAction(globalCfg), root, "", "Opening")
		p := tea.NewProgram(model, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		if m, ok := finalModel.(ui.PickerModel); ok {
			if path := m.SelectedPath(); path != "" {
				bin, args := globalCfg.EditorArgs(path)
				cmd := exec.Command(bin, args...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to open editor: %w", err)
				}
			}
		}

		return nil
	},
}
