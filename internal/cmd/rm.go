package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/krystof/worktree-workflow/internal/config"
	"github.com/krystof/worktree-workflow/internal/git"
	"github.com/krystof/worktree-workflow/internal/ui"
)

var forceRemove bool

var rmCmd = &cobra.Command{
	Use:     "rm [branch]",
	Aliases: []string{"remove"},
	Short:   "Remove a worktree, prune, and clean up",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := git.Root()
		if err != nil {
			return fmt.Errorf("not a git repository")
		}

		globalForce, _ := cmd.Root().PersistentFlags().GetBool("force")
		effectiveForce := forceRemove || globalForce

		globalCfg, err := config.LoadGlobal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load global config: %s\n", err)
		}

		// If branch given, remove directly
		if len(args) == 1 {
			branch := args[0]
			worktreeDir := git.WorktreeDir(root, globalCfg.Naming.WorktreeDirSuffix, globalCfg.Naming.BranchSeparator, branch)
			return removeWorktree(worktreeDir, branch, effectiveForce)
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

		model := ui.NewPickerModel("Remove worktree", worktrees, ui.RemoveWorktreeAction(effectiveForce), root, root)
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

func init() {
	rmCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Force remove even with uncommitted changes")
}

func removeWorktree(worktreeDir, branch string, force bool) error {
	out, err := git.WorktreeRemove(worktreeDir, force)
	if err != nil {
		// Prompt for force if modified/untracked files (unless force already set)
		if !force && strings.Contains(err.Error(), "modified or untracked") {
			fmt.Printf("! Worktree '%s' contains modified or untracked files.\n", branch)
			fmt.Print("  Force remove? [y/N] ")

			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer == "y" {
				if _, err = git.WorktreeRemove(worktreeDir, true); err != nil {
					return err
				}
				fmt.Printf("✓ Force removed worktree at %s\n", worktreeDir)
			} else {
				fmt.Println("Aborted.")
				return nil
			}
		} else {
			return err
		}
	} else {
		if out != "" {
			fmt.Print(out)
		}
		fmt.Printf("✓ Removed worktree at %s\n", worktreeDir)
	}

	if out, err := git.WorktreePrune(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: git worktree prune failed: %s\n", err)
	} else if out != "" {
		fmt.Print(out)
	}
	fmt.Println("✓ Pruned stale worktree references")
	return nil
}
