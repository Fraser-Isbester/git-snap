package commands

import (
	"fmt"

	"github.com/fraser-isbester/git-snap/pkg/config"
	"github.com/spf13/cobra"
)

func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize git-snap in the current directory",
		Long: `Initialize git-snap by creating default configuration files and
setting up the necessary directory structure.`,
		RunE: runInit,
	}

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := getConfigPath()

	fmt.Printf("Initializing git-snap...\n")
	fmt.Printf("Configuration directory: %s\n", configPath)

	configManager := config.NewConfigManager(configPath)

	if err := configManager.CreateDefaultConfigs(); err != nil {
		return fmt.Errorf("failed to create default configurations: %w", err)
	}

	fmt.Println("✓ Default configurations created")
	fmt.Println("✓ git-snap initialized successfully")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Run 'git-snap config list' to see available configurations")
	fmt.Println("2. Run 'git-snap correlate --help' to see correlation options")
	fmt.Println("3. Prepare your events file (JSON, JSONL, or CSV format)")
	fmt.Println()
	fmt.Println("Example usage:")
	fmt.Println("  git-snap correlate -e events.json -c ai-inference")

	return nil
}
