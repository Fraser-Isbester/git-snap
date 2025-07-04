package main

import (
	"fmt"
	"os"

	"github.com/fraser-isbester/git-snap/cmd/git-snap/commands"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "git-snap",
		Short: "Git Correlation Engine - Snap events to git commits",
		Long: `Git Snap is a tool for correlating timestamped events with git commits
based on configurable attribute matching and temporal proximity.

It enables automated tagging and analysis of commits based on external
system events like AI inference calls, deployments, build events, etc.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.AddCommand(commands.NewCorrelateCommand())
	rootCmd.AddCommand(commands.NewConfigCommand())
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewVersionCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
