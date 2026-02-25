// Package cmd provides commands for gitlab-issue-report.
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultAPITimeout = 30 * time.Second // Default timeout for GitLab API requests
)

// commandOptions holds all CLI flag values for issue commands.
type commandOptions struct {
	logLevel      string        // Log level: info, warn, error, debug
	projectIDFlag int64         // Project ID
	groupIDFlag   int64         // Group ID
	createdFilter bool          // Filter by created date
	updatedFilter bool          // Filter by updated date
	stateFilter   string        // Filter by state: "opened", "closed", "all"
	formatOutput  string        // Output format: "plain", "table", "markdown"
	debugFlag     bool          // Shorthand for debug logging
	verboseFlag   bool          // Shorthand for verbose logging
	interval      string        // Date interval
	mineOption    bool          // Filter issues assigned to current user
	apiTimeout    time.Duration // API request timeout
	timezone      string        // Timezone for date calculations
}

// opts is the package-level command options instance for Cobra flag binding.
var opts commandOptions

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gitlab-issue-report",
	Short: "Report and filter GitLab issues from projects and groups",
	Long: `gitlab-issue-report - A CLI tool to fetch and display GitLab issues.

Retrieves issues from GitLab projects or groups with filtering options.
Supports multiple output formats and can auto-detect your project from
the current git repository.

AUTHENTICATION:
  Set the GITLAB_TOKEN environment variable with your GitLab personal access token.
  Optionally set GITLAB_URI for self-hosted instances (defaults to https://gitlab.com).
  Optionally set GITLAB_API_TIMEOUT for custom API timeout (e.g., "1m").
  Optionally set GITLAB_TIMEZONE for date calculation timezone (e.g., "America/New_York").

EXAMPLES:
  # Auto-detect project from current git repository
  gitlab-issue-report project

  # Specify a project ID explicitly
  gitlab-issue-report project -p 12345

  # Get issues from a group
  gitlab-issue-report group -g 678

  # Filter by date interval (last 7 days)
  gitlab-issue-report project -i "/-7/ ::"

  # Get only closed issues
  gitlab-issue-report project --state closed

  # Output as a markdown table
  gitlab-issue-report project --format markdown

  # Combine filters: closed issues from last 30 days
  gitlab-issue-report project -p 12345 -i "/-30/ ::" --state closed --format table

  # Show only issues assigned to you
  gitlab-issue-report project --mine

  # Use a specific timezone for date calculations
  gitlab-issue-report project -i "/-7/ ::" --timezone "America/New_York"

For more details on each subcommand:
  gitlab-issue-report project --help
  gitlab-issue-report group --help`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}
	return nil
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// ===== PERSISTENT FLAGS (ALL COMMANDS) =====
	rootCmd.PersistentFlags().DurationVar(&opts.apiTimeout, "api-timeout", defaultAPITimeout,
		"Timeout for GitLab API requests (e.g., 30s, 1m)")
	rootCmd.PersistentFlags().StringVarP(&opts.timezone, "timezone", "T", "",
		"Timezone for date calculations (e.g., America/New_York, UTC, Local)")

	// ===== PROJECT COMMAND FLAGS =====

	// Project command flags
	projectCmd.Flags().StringVarP(&opts.interval, "interval", "i", "", "Date interval (e.g., '/-1/ ::' for last month)")
	projectCmd.Flags().StringVar(&opts.logLevel, "log-level", "error", "Log level: info, warn, error, debug")
	projectCmd.Flags().BoolVarP(&opts.debugFlag, "debug", "d", false,
		"Enable debug logging (shorthand for --log-level=debug)")
	projectCmd.Flags().BoolVarP(&opts.verboseFlag, "verbose", "v", false,
		"Enable verbose logging (shorthand for --log-level=info)")

	projectCmd.Flags().Int64Var(&opts.projectIDFlag, "project-id", 0,
		"Project ID to get issues from (auto-detected from git if not set)")
	projectCmd.Flags().Int64VarP(&opts.projectIDFlag, "project", "p", 0, "Project ID (alias for --project-id)")

	projectCmd.Flags().BoolVar(&opts.createdFilter, "created", false,
		"Filter issues by creation date (requires --interval)")
	projectCmd.Flags().BoolVarP(&opts.updatedFilter, "updated", "U", false,
		"Filter issues by update date (requires --interval)")

	projectCmd.Flags().StringVar(&opts.stateFilter, "state", "", "Filter by state: opened, closed, all")
	projectCmd.Flags().StringVar(&opts.formatOutput, "format", "plain", "Output format: plain, table, markdown")

	projectCmd.Flags().BoolVarP(&opts.mineOption, "mine", "M", false, "Only issues assigned to current user")

	rootCmd.AddCommand(projectCmd)

	// ===== GROUP COMMAND FLAGS =====

	// Group command flags
	groupCmd.Flags().StringVarP(&opts.interval, "interval", "i", "", "Date interval (e.g., '/-1/ ::' for last month)")
	groupCmd.Flags().StringVar(&opts.logLevel, "log-level", "error", "Log level: info, warn, error, debug")
	groupCmd.Flags().BoolVarP(&opts.debugFlag, "debug", "d", false,
		"Enable debug logging (shorthand for --log-level=debug)")
	groupCmd.Flags().BoolVarP(&opts.verboseFlag, "verbose", "v", false,
		"Enable verbose logging (shorthand for --log-level=info)")

	groupCmd.Flags().Int64Var(&opts.groupIDFlag, "group-id", 0, "Group ID to get issues from (required)")
	groupCmd.Flags().Int64VarP(&opts.groupIDFlag, "group", "g", 0, "Group ID (alias for --group-id)")

	groupCmd.Flags().BoolVar(&opts.createdFilter, "created", false, "Filter issues by creation date (requires --interval)")
	groupCmd.Flags().BoolVarP(&opts.updatedFilter, "updated", "U", false,
		"Filter issues by update date (requires --interval)")

	groupCmd.Flags().StringVar(&opts.stateFilter, "state", "", "Filter by state: opened, closed, all")
	groupCmd.Flags().StringVar(&opts.formatOutput, "format", "plain", "Output format: plain, table, markdown")

	groupCmd.Flags().BoolVarP(&opts.mineOption, "mine", "M", false, "Only issues assigned to current user")

	rootCmd.AddCommand(groupCmd)
}
