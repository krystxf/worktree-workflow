package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/krystof/worktree-workflow/internal/config"
	"github.com/krystof/worktree-workflow/internal/git"
	"github.com/krystof/worktree-workflow/internal/ui"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List worktrees and select one to open",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := git.Root()
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

		model := ui.NewPickerModel("Worktrees", worktrees, ui.OpenEditorAction(globalCfg), "")
		p := tea.NewProgram(model, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		if msg := finalModel.(ui.PickerModel).ResultMessage(); msg != "" {
			fmt.Print(msg)
		}

		return nil
	},
}
