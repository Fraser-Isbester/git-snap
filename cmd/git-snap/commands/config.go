package commands

import (
	"fmt"

	"github.com/fraser-isbester/git-snap/pkg/config"
	"github.com/spf13/cobra"
)

func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage correlation configurations",
		Long:  `Create, list, and manage correlation configurations.`,
	}

	cmd.AddCommand(newConfigListCommand())
	cmd.AddCommand(newConfigInitCommand())
	cmd.AddCommand(newConfigShowCommand())

	return cmd
}

func newConfigListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager := config.NewConfigManager(getConfigPath())
			configs, err := configManager.ListConfigs()
			if err != nil {
				return fmt.Errorf("failed to list configurations: %w", err)
			}

			if len(configs) == 0 {
				fmt.Println("No configurations found. Run 'git-snap config init' to create default configurations.")
				return nil
			}

			fmt.Println("Available configurations:")
			for _, configName := range configs {
				fmt.Printf("  - %s\n", configName)
			}

			return nil
		},
	}
}

func newConfigInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize default configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager := config.NewConfigManager(getConfigPath())
			if err := configManager.CreateDefaultConfigs(); err != nil {
				return fmt.Errorf("failed to create default configurations: %w", err)
			}

			fmt.Printf("Default configurations created in %s\n", getConfigPath())
			fmt.Println("Available configurations:")
			fmt.Println("  - default: Basic correlation configuration")
			fmt.Println("  - ai-inference: Configuration for AI inference events")
			fmt.Println("  - deployment: Configuration for deployment events")

			return nil
		},
	}
}

func newConfigShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <config-name>",
		Short: "Show configuration details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configName := args[0]
			configManager := config.NewConfigManager(getConfigPath())

			snapConfig, err := configManager.LoadConfig(configName)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			fmt.Printf("Configuration: %s\n", configName)
			fmt.Printf("Time Window: %s\n", snapConfig.TimeWindow)
			fmt.Printf("Score Weights:\n")
			if len(snapConfig.ScoreWeights) == 0 {
				fmt.Printf("  (none)\n")
			} else {
				for key, weight := range snapConfig.ScoreWeights {
					fmt.Printf("  %s: %.2f\n", key, weight)
				}
			}
			fmt.Printf("Attribute Rules:\n")
			if len(snapConfig.AttributeRules) == 0 {
				fmt.Printf("  (none)\n")
			} else {
				for i, rule := range snapConfig.AttributeRules {
					fmt.Printf("  %d. %s -> %s (%s)%s\n",
						i+1, rule.EventKey, rule.CommitKey, rule.MatchType.String(),
						func() string {
							if rule.Required {
								return " [required]"
							}
							return ""
						}())
				}
			}

			return nil
		},
	}

	return cmd
}
