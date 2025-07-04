package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fraser-isbester/git-snap/pkg/config"
	"github.com/fraser-isbester/git-snap/pkg/correlation"
	"github.com/fraser-isbester/git-snap/pkg/git"
	"github.com/fraser-isbester/git-snap/pkg/parser"
	"github.com/fraser-isbester/git-snap/pkg/types"
	"github.com/spf13/cobra"
)

func NewCorrelateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "correlate",
		Short: "Correlate events with git commits",
		Long: `Correlate timestamped events with git commits based on configurable
attribute matching and temporal proximity.`,
		RunE: runCorrelate,
	}

	cmd.Flags().StringP("events", "e", "", "Path to events file (JSON, JSONL, or CSV)")
	cmd.Flags().StringP("config", "c", "default", "Configuration name to use")
	cmd.Flags().StringP("repo", "r", ".", "Path to git repository")
	cmd.Flags().StringP("since", "s", "7d", "Time window to look back for commits (e.g., 7d, 24h, 30m)")
	cmd.Flags().StringP("output", "o", "json", "Output format (json, table)")
	cmd.Flags().Float64P("threshold", "t", 0.5, "Minimum correlation score threshold")
	cmd.Flags().BoolP("verbose", "v", false, "Verbose output")

	cmd.MarkFlagRequired("events")

	return cmd
}

func runCorrelate(cmd *cobra.Command, args []string) error {
	eventsFile, _ := cmd.Flags().GetString("events")
	configName, _ := cmd.Flags().GetString("config")
	repoPath, _ := cmd.Flags().GetString("repo")
	sinceStr, _ := cmd.Flags().GetString("since")
	outputFormat, _ := cmd.Flags().GetString("output")
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Loading events from: %s\n", eventsFile)
		fmt.Printf("Using configuration: %s\n", configName)
		fmt.Printf("Repository path: %s\n", repoPath)
	}

	events, err := parser.ParseEventsFromFile(eventsFile)
	if err != nil {
		return fmt.Errorf("failed to parse events: %w", err)
	}

	if verbose {
		fmt.Printf("Loaded %d events\n", len(events))
	}

	repoPath, err = git.FindGitRepository(repoPath)
	if err != nil {
		return fmt.Errorf("failed to find git repository: %w", err)
	}

	gitClient := git.NewGitClient(repoPath)

	since, err := parseTimeWindow(sinceStr)
	if err != nil {
		return fmt.Errorf("invalid time window: %w", err)
	}

	commits, err := gitClient.GetCommits(since)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if verbose {
		fmt.Printf("Found %d commits since %s\n", len(commits), since.Format("2006-01-02"))
	}

	configManager := config.NewConfigManager(getConfigPath())
	snapConfig, err := configManager.LoadConfig(configName)
	if err != nil {
		if verbose {
			fmt.Printf("Failed to load config %s, using default: %v\n", configName, err)
		}
		snapConfig = config.DefaultConfig()
		if verbose {
			fmt.Printf("Default config created with %d rules\n", len(snapConfig.AttributeRules))
		}
	} else {
		if verbose {
			fmt.Printf("Successfully loaded config from file\n")
		}
	}

	// TEMPORARY: Force use of hardcoded config for demo
	if len(snapConfig.AttributeRules) == 0 {
		if verbose {
			fmt.Printf("Config has no rules, using hardcoded default\n")
		}
		snapConfig = config.DefaultConfig()
	}

	if verbose {
		fmt.Printf("Configuration loaded - Rules: %d, TimeWindow: %s\n", len(snapConfig.AttributeRules), snapConfig.TimeWindow)
		for i, rule := range snapConfig.AttributeRules {
			fmt.Printf("  Rule %d: %s -> %s (%s, required: %t)\n", i+1, rule.EventKey, rule.CommitKey, rule.MatchType.String(), rule.Required)
		}
	}

	engine := correlation.NewCorrelationEngine(*snapConfig)
	results := engine.SnapToCommits(events, commits)

	filteredResults := make([]types.CorrelationResult, 0)
	for _, result := range results {
		if result.Score >= threshold {
			filteredResults = append(filteredResults, result)
		}
	}

	if verbose {
		fmt.Printf("Found %d correlations above threshold %.2f\n", len(filteredResults), threshold)
	}

	return outputResults(filteredResults, outputFormat)
}

func parseTimeWindow(window string) (time.Time, error) {
	// Handle days manually since Go's time.ParseDuration doesn't support days
	if strings.HasSuffix(window, "d") {
		daysStr := strings.TrimSuffix(window, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid duration format: %s", window)
		}
		duration := time.Duration(days) * 24 * time.Hour
		return time.Now().Add(-duration), nil
	}

	duration, err := time.ParseDuration(window)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration format: %s", window)
	}
	return time.Now().Add(-duration), nil
}

func outputResults(results []types.CorrelationResult, format string) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(results)
	case "table":
		return outputTable(results)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func outputTable(results []types.CorrelationResult) error {
	fmt.Printf("%-8s %-12s %-40s %-10s %-15s\n", "Score", "Event ID", "Commit SHA", "Time Delta", "Author")
	fmt.Println(strings.Repeat("-", 90))

	for _, result := range results {
		fmt.Printf("%.3f    %-12s %-40s %-10s %-15s\n",
			result.Score,
			result.Event.ID,
			result.Commit.SHA[:8],
			formatDuration(result.TimeDelta),
			result.Commit.Author)
	}

	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
