package cmd

import (
	"bufio"
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

		globalForce, _ := cmd.Root().PersistentFlags().GetBool("force")
		isTerminal := term.IsTerminal(int(os.Stdin.Fd()))

		createNewBranch := false
		if !git.BranchExists(branch) {
			if globalForce {
				createNewBranch = true
			} else if !isTerminal {
				return fmt.Errorf("branch '%s' does not exist (create it first or run interactively)", branch)
			} else {
				fmt.Printf("Branch '%s' does not exist. Create it? [y/N] ", branch)
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println("Aborted.")
					return nil
				}
				createNewBranch = true
			}
		}

		if !isTerminal {
			return createPlain(branch, root, worktreeDir, globalCfg, projectCfg, createNewBranch)
		}

		model := ui.NewCreateModel(branch, root, worktreeDir, globalCfg, projectCfg, createNewBranch)
		p := tea.NewProgram(model)

		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		if m, ok := finalModel.(ui.CreateModel); ok && m.Err() != nil {
			return m.Err()
		}

		return nil
	},
}

func createPlain(branch, root, worktreeDir string, globalCfg config.GlobalConfig, projectCfg config.ProjectConfig, createNewBranch bool) error {
	if createNewBranch {
		fmt.Printf("Creating branch '%s'...\n", branch)
		out, err := git.BranchCreate(branch)
		if out != "" {
			fmt.Print(out)
		}
		if err != nil {
			return err
		}
	}

	fmt.Printf("Creating worktree for '%s'...\n", branch)
	out, err := git.WorktreeAdd(worktreeDir, branch)
	if out != "" {
		fmt.Print(out)
	}
	if err != nil {
		return err
	}

	if projectCfg.SyncIgnored != nil && *projectCfg.SyncIgnored {
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
			errMsg := strings.TrimSpace(string(out))
			if errMsg != "" {
				return fmt.Errorf("hook %q failed: %w: %s", hook, err, errMsg)
			}
			return fmt.Errorf("hook %q failed: %w", hook, err)
		}
	}

	if globalCfg.AutoOpenEditor != nil && *globalCfg.AutoOpenEditor {
		fmt.Printf("Opening in %s...\n", globalCfg.Editor)
		bin, editorArgs := globalCfg.EditorArgs(worktreeDir)
		cmd := exec.Command(bin, editorArgs...)
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
