package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/krystof/worktree-workflow/internal/config"
	"github.com/krystof/worktree-workflow/internal/git"
	syncpkg "github.com/krystof/worktree-workflow/internal/sync"
	"github.com/krystof/worktree-workflow/internal/ui"
)

var createCmd = &cobra.Command{
	Use:   "create <branch>",
	Short: "Create a new worktree for the given branch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]

		root, err := git.Root()
		if err != nil {
			return fmt.Errorf("not a git repository")
		}

		globalCfg, err := config.LoadGlobal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load global config: %s\n", err)
		}

		projectCfg, err := config.LoadProject(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load project config: %s\n", err)
		}

		worktreeDir := git.WorktreeDir(root, globalCfg.Naming.WorktreeDirSuffix, globalCfg.Naming.BranchSeparator, branch)

		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return createPlain(branch, root, worktreeDir, globalCfg, projectCfg)
		}

		model := ui.NewCreateModel(branch, root, worktreeDir, globalCfg, projectCfg)
		p := tea.NewProgram(model)

		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

func createPlain(branch, root, worktreeDir string, globalCfg config.GlobalConfig, projectCfg config.ProjectConfig) error {
	fmt.Printf("Creating worktree for '%s'...\n", branch)
	out, err := git.WorktreeAdd(worktreeDir, branch)
	if out != "" {
		fmt.Print(out)
	}
	if err != nil {
		return err
	}

	if projectCfg.SyncIgnored {
		fmt.Println("Syncing gitignored files...")
		out, err := syncpkg.SyncIgnored(root, worktreeDir, projectCfg.SyncExcludes)
		if out != "" {
			fmt.Print(out)
		}
		if err != nil {
			return err
		}
	}

	for i, hook := range projectCfg.PostCopyHooks {
		fmt.Printf("Running hook [%d/%d]: %s\n", i+1, len(projectCfg.PostCopyHooks), hook)
		cmd := exec.Command("sh", "-c", hook)
		cmd.Dir = worktreeDir
		out, err := cmd.CombinedOutput()
		if len(out) > 0 {
			fmt.Print(string(out))
		}
		if err != nil {
			return fmt.Errorf("hook %q failed: %s", hook, strings.TrimSpace(string(out)))
		}
	}

	if globalCfg.AutoOpenEditor {
		fmt.Printf("Opening in %s...\n", globalCfg.Editor)
		cmd := exec.Command(globalCfg.Editor, worktreeDir)
		out, err := cmd.CombinedOutput()
		if len(out) > 0 {
			fmt.Print(string(out))
		}
		if err != nil {
			return fmt.Errorf("failed to open editor: %s", strings.TrimSpace(string(out)))
		}
	}

	fmt.Printf("Done! %s\n", worktreeDir)
	return nil
}
