package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "wtw",
	Short:   "Worktree workflow manager",
	Long:    "A TUI tool for managing git worktrees — create, list, and remove worktrees with ease.",
	Version: version,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		clean, _ := cmd.Flags().GetBool("clean")
		if clean {
			return cleanWorktrees(cmd)
		}
		return cmd.Help()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("wtw does not support Windows")
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().Bool("clean", false, "Remove clean worktrees that exactly match their remote upstream (never pushes)")
	rootCmd.PersistentFlags().BoolP("force", "f", false, "Do not prompt (force remove dirty worktrees, create branch when missing)")
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(openCmd)
}
