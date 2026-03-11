package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wtw",
	Short: "Worktree workflow manager",
	Long:  "A TUI tool for managing git worktrees — create, list, and remove worktrees with ease.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolP("force", "f", false, "Do not prompt (force remove dirty worktrees, create branch when missing)")
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(rmCmd)
}
