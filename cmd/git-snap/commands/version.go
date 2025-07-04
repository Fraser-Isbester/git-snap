package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("git-snap version %s\n", Version)
		},
	}
}
