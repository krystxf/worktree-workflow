package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wtw",
	Short: "Worktree workflow manager",
	Long:  "A TUI tool for managing git worktrees — create, list, and remove worktrees with ease.",
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
	rootCmd.PersistentFlags().BoolP("force", "f", false, "Do not prompt (force remove dirty worktrees, create branch when missing)")
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(rmCmd)
}
