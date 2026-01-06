package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var closedOption bool
var openedOption bool
var createdAtOption bool
var updatedAtOption bool
var interval string
var projectID int64
var groupID int64
var debugLevel string
var markdownOutput bool
var mineOption bool

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
	projectCmd.Flags().StringVarP(&interval, "i", "i", "", "interval, ex '/-1/ ::' to describe the interval of last month")
	projectCmd.Flags().StringVarP(&debugLevel, "d", "d", "error", "Debug level (info,warn,debug)")
	projectCmd.Flags().BoolVarP(&closedOption, "closed", "c", false, "only closed issues")
	projectCmd.Flags().BoolVarP(&openedOption, "opened", "o", false, "only opened issues")
	projectCmd.Flags().BoolVarP(&createdAtOption, "createdAt", "r", false, "issues filtered with created date")
	projectCmd.Flags().BoolVarP(&updatedAtOption, "updatedAt", "u", false, "issues filtered with updated date")
	projectCmd.Flags().Int64VarP(&projectID, "id", "p", 0, "Project ID to get issues from")
	projectCmd.Flags().BoolVarP(&markdownOutput, "markdown", "m", false, "output in markdown format")
	projectCmd.Flags().BoolVarP(&mineOption, "mine", "M", false, "only issues assigned to current user")
	rootCmd.AddCommand(projectCmd)

	groupCmd.Flags().StringVarP(&interval, "i", "i", "", "interval, ex '/-1/ ::' to describe the interval of last month")
	groupCmd.Flags().StringVarP(&debugLevel, "d", "d", "error", "Debug level (info,warn,debug)")
	groupCmd.Flags().BoolVarP(&closedOption, "closed", "c", false, "only closed issues")
	groupCmd.Flags().BoolVarP(&openedOption, "opened", "o", false, "only opened issues")
	groupCmd.Flags().BoolVarP(&createdAtOption, "createdAt", "r", false, "issues filtered with created date")
	groupCmd.Flags().BoolVarP(&updatedAtOption, "updatedAt", "u", false, "issues filtered with updated date")
	groupCmd.Flags().Int64VarP(&groupID, "id", "g", 0, "Group ID to get issues from")
	groupCmd.Flags().BoolVarP(&markdownOutput, "markdown", "m", false, "output in markdown format")
	groupCmd.Flags().BoolVarP(&mineOption, "mine", "M", false, "only issues assigned to current user")
	rootCmd.AddCommand(groupCmd)
}
