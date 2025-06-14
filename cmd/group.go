package cmd

import (
	"os"

	"github.com/sgaunet/gitlab-issue-report/internal/core"
	"github.com/sgaunet/gitlab-issue-report/internal/render"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// groupCmd represents the group command
var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Get issues of a GitLab group",
	Long:  `Get issues of a GitLab group by ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if group ID is provided
		if groupID == 0 {
			logrus.Errorln("Group ID is required. Please provide it with the --id flag.")
			cmd.Help()
			os.Exit(1)
		}

		// Initialize logging
		initTrace(debugLevel)

		// Setup environment
		setupEnvironment()

		// Parse interval if provided
		beginTime, endTime := parseInterval(interval)

		// Create GitLab client
		app, err := core.NewApp(os.Getenv("GITLAB_TOKEN"), os.Getenv("GITLAB_URI"))
		if err != nil {
			logrus.Errorln(err.Error())
			os.Exit(1)
		}

		// Build issue retrieval options
		options := buildIssueOptions(0, groupID, beginTime, endTime)

		// Get and display issues
		issues, err := app.GetIssues(options...)
		if err != nil {
			logrus.Errorln(err.Error())
			os.Exit(1)
		}

		render.PrintIssues(issues, true)
	},
}
