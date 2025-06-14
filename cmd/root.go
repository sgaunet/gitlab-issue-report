package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var closedOption bool
var openedOption bool
var createdAtOption bool
var updatedAtOption bool
var interval string
var projectID int
var groupID int
var debugLevel string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gitlab-issue-report",
	Short: "Tool to get issues of a gitlab project or group.",
	Long: `Tool to get issues of a gitlab project or group. 
	`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	projectCmd.Flags().StringVarP(&interval, "i", "i", "", "interval, ex '/-1/ ::' to describe the interval of last month")
	projectCmd.Flags().StringVarP(&debugLevel, "d", "d", "error", "Debug level (info,warn,debug)")
	projectCmd.Flags().BoolVarP(&closedOption, "closed", "c", false, "only closed issues")
	projectCmd.Flags().BoolVarP(&openedOption, "opened", "o", false, "only opened issues")
	projectCmd.Flags().BoolVarP(&createdAtOption, "createdAt", "r", false, "issues filtered with created date")
	projectCmd.Flags().BoolVarP(&updatedAtOption, "updatedAt", "u", false, "issues filtered with updated date")
	projectCmd.Flags().IntVarP(&projectID, "id", "p", 0, "Project ID to get issues from")
	rootCmd.AddCommand(projectCmd)

	groupCmd.Flags().StringVarP(&interval, "i", "i", "", "interval, ex '/-1/ ::' to describe the interval of last month")
	groupCmd.Flags().StringVarP(&debugLevel, "d", "d", "error", "Debug level (info,warn,debug)")
	groupCmd.Flags().BoolVarP(&closedOption, "closed", "c", false, "only closed issues")
	groupCmd.Flags().BoolVarP(&openedOption, "opened", "o", false, "only opened issues")
	groupCmd.Flags().BoolVarP(&createdAtOption, "createdAt", "r", false, "issues filtered with created date")
	groupCmd.Flags().BoolVarP(&updatedAtOption, "updatedAt", "u", false, "issues filtered with updated date")
	groupCmd.Flags().IntVarP(&groupID, "id", "g", 0, "Group ID to get issues from")
	rootCmd.AddCommand(groupCmd)
}
