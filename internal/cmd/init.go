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

var initLocal bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize global config (~/.config/worktree-workflow) or project config (--local)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if initLocal {
			return runInitLocal(cmd)
		}
		return runInitGlobal()
	},
}

func init() {
	initCmd.Flags().BoolVar(&initLocal, "local", false, "Initialize project config (.worktree-workflow.json) in the current repo")
	rootCmd.AddCommand(initCmd)
}

func runInitGlobal() error {
	globalCfg, err := config.LoadGlobal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load existing config: %s\n", err)
	}

	model := ui.NewInitModel(globalCfg)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	m, ok := finalModel.(ui.InitModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}
	result := m.Result()
	if result.Canceled {
		fmt.Println("Canceled.")
		return nil
	}

	globalCfg.Editor = result.Editor
	globalCfg.Naming.WorktreeDirSuffix = result.DirSuffix
	globalCfg.Naming.BranchSeparator = result.BranchSep
	if err := writeGlobalConfig(globalCfg); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}
	fmt.Printf("\n  ✓ Global config written to ~/.config/worktree-workflow/config.json\n\n")

	return nil
}

func runInitLocal(cmd *cobra.Command) error {
	root, err := git.Root()
	if err != nil {
		return fmt.Errorf("not a git repository (--local requires a git repo)")
	}

	globalForce, _ := cmd.Root().PersistentFlags().GetBool("force")
	if globalForce {
		projectCfg := config.DefaultProject()
		projectPath := filepath.Join(root, ".worktree-workflow.json")
		if err := writeProjectConfigData(projectPath, projectCfg); err != nil {
			return fmt.Errorf("failed to write project config: %w", err)
		}
		fmt.Printf("\n  ✓ Project config written to %s\n\n", projectPath)
		return nil
	}

	projectCfg, err := config.LoadProject(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load existing project config: %s\n", err)
	}

	model := ui.NewLocalInitModel(projectCfg)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	lm, ok := finalModel.(ui.LocalInitModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}
	result := lm.Result()
	if result.Canceled {
		fmt.Println("Canceled.")
		return nil
	}

	v := result.SyncIgnored
	projectCfg.SyncIgnored = &v
	projectCfg.SyncExcludes = result.SyncExcludes
	projectCfg.PostCopyHooks = result.PostCopyHooks

	projectPath := filepath.Join(root, ".worktree-workflow.json")
	if err := writeProjectConfigData(projectPath, projectCfg); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}
	fmt.Printf("\n  ✓ Project config written to %s\n\n", projectPath)

	return nil
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

func writeProjectConfigData(path string, cfg config.ProjectConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0o644)
}
