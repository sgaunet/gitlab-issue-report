package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// CLI flag variables
var (
	logLevel      string // Log level: info, warn, error, debug
	projectIDFlag int64  // Project ID
	groupIDFlag   int64  // Group ID
	createdFilter bool   // Filter by created date
	updatedFilter bool   // Filter by updated date
	stateFilter   string // Filter by state: "opened", "closed", "all"
	formatOutput  string // Output format: "plain", "table", "markdown"
	debugFlag     bool   // Shorthand for debug logging
	verboseFlag   bool   // Shorthand for verbose logging
	interval      string // Date interval
	mineOption    bool   // Filter issues assigned to current user
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gitlab-issue-report",
	Short: "Tool to get issues of a gitlab project or group.",
	Long: `Tool to get issues of a gitlab project or group. 
	`,
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

	// ===== PROJECT COMMAND FLAGS =====

	// Project command flags
	projectCmd.Flags().StringVarP(&interval, "interval", "i", "", "Date interval (e.g., '/-1/ ::' for last month)")
	projectCmd.Flags().StringVar(&logLevel, "log-level", "error", "Log level: info, warn, error, debug")
	projectCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debug logging (shorthand for --log-level=debug)")
	projectCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose logging (shorthand for --log-level=info)")

	projectCmd.Flags().Int64Var(&projectIDFlag, "project-id", 0, "Project ID to get issues from (auto-detected from git if not set)")
	projectCmd.Flags().Int64VarP(&projectIDFlag, "project", "p", 0, "Project ID (alias for --project-id)")

	projectCmd.Flags().BoolVar(&createdFilter, "created", false, "Filter issues by creation date (requires --interval)")
	projectCmd.Flags().BoolVarP(&updatedFilter, "updated", "U", false, "Filter issues by update date (requires --interval)")

	projectCmd.Flags().StringVar(&stateFilter, "state", "", "Filter by state: opened, closed, all")
	projectCmd.Flags().StringVar(&formatOutput, "format", "plain", "Output format: plain, table, markdown")

	projectCmd.Flags().BoolVarP(&mineOption, "mine", "M", false, "Only issues assigned to current user")

	rootCmd.AddCommand(projectCmd)

	// ===== GROUP COMMAND FLAGS =====

	// Group command flags
	groupCmd.Flags().StringVarP(&interval, "interval", "i", "", "Date interval (e.g., '/-1/ ::' for last month)")
	groupCmd.Flags().StringVar(&logLevel, "log-level", "error", "Log level: info, warn, error, debug")
	groupCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debug logging (shorthand for --log-level=debug)")
	groupCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose logging (shorthand for --log-level=info)")

	groupCmd.Flags().Int64Var(&groupIDFlag, "group-id", 0, "Group ID to get issues from (required)")
	groupCmd.Flags().Int64VarP(&groupIDFlag, "group", "g", 0, "Group ID (alias for --group-id)")

	groupCmd.Flags().BoolVar(&createdFilter, "created", false, "Filter issues by creation date (requires --interval)")
	groupCmd.Flags().BoolVarP(&updatedFilter, "updated", "U", false, "Filter issues by update date (requires --interval)")

	groupCmd.Flags().StringVar(&stateFilter, "state", "", "Filter by state: opened, closed, all")
	groupCmd.Flags().StringVar(&formatOutput, "format", "plain", "Output format: plain, table, markdown")

	groupCmd.Flags().BoolVarP(&mineOption, "mine", "M", false, "Only issues assigned to current user")

	rootCmd.AddCommand(groupCmd)
}
