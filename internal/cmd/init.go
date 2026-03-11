package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/krystof/worktree-workflow/internal/config"
	"github.com/krystof/worktree-workflow/internal/git"
	"github.com/krystof/worktree-workflow/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize global config (~/.config/worktree-workflow) and optionally project config",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalCfg, _ := config.LoadGlobal()

		model := ui.NewInitModel(globalCfg.Editor)
		p := tea.NewProgram(model)

		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		result := finalModel.(ui.InitModel).Result()
		if result.Canceled {
			fmt.Println("Canceled.")
			return nil
		}

		// Write global config
		globalCfg.Editor = result.Editor
		if err := writeGlobalConfig(globalCfg); err != nil {
			return fmt.Errorf("failed to write global config: %w", err)
		}
		fmt.Printf("\n  ✓ Global config written to ~/.config/worktree-workflow/config.json\n")

		// Write project config if inside a git repo
		if root, err := git.Root(); err == nil {
			projectPath := filepath.Join(root, ".worktree-workflow.json")
			if err := writeProjectConfig(projectPath); err != nil {
				return fmt.Errorf("failed to write project config: %w", err)
			}
			fmt.Printf("  ✓ Project config written to %s\n", projectPath)
		}

		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func writeGlobalConfig(cfg config.GlobalConfig) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".config", "worktree-workflow")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "config.json"), append(data, '\n'), 0o644)
}

func writeProjectConfig(path string) error {
	cfg := config.DefaultProject()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0o644)
}
