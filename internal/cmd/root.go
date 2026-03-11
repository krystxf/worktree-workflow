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
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(psCmd)
	rootCmd.AddCommand(rmCmd)
}
