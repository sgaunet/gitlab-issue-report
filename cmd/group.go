package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/sgaunet/gitlab-issue-report/internal/core"
	"github.com/sgaunet/gitlab-issue-report/internal/render"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	errGroupIDRequired = errors.New("group ID is required")
)

// groupCmd represents the group command.
var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Get issues of a GitLab group",
	Long:  `Get issues of a GitLab group by ID.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Reconcile old and new flags
		if err := reconcileFlags(cmd); err != nil {
			return err
		}

		// Check if group ID is provided
		if groupIDFlag == 0 {
			logrus.Errorln("Group ID is required. Please provide it with the --group-id or --group flag.")
			if err := cmd.Help(); err != nil {
				logrus.Errorln("Failed to display help:", err)
			}
			return errGroupIDRequired
		}

		// Initialize logging with new log level variable
		initTrace(logLevel)

		// Setup environment
		if err := setupEnvironment(); err != nil {
			logrus.Errorln(err.Error())
			return err
		}

		// Apply timeout from environment variable if flag not set
		applyTimeoutFromEnv(cmd.Flags().Changed("api-timeout"))

		// Parse interval if provided
		beginTime, endTime, err := parseInterval(interval)
		if err != nil {
			logrus.Errorln(err.Error())
			return err
		}

		// Create GitLab client
		app, err := core.NewApp(os.Getenv("GITLAB_TOKEN"), os.Getenv("GITLAB_URI"), apiTimeout)
		if err != nil {
			logrus.Errorln(err.Error())
			return fmt.Errorf("failed to create GitLab client: %w", err)
		}

		// Build issue retrieval options
		options, err := buildIssueOptions(0, groupIDFlag, beginTime, endTime)
		if err != nil {
			logrus.Errorln(err.Error())
			return err
		}

		// Get and display issues
		issues, err := app.GetIssues(options...)
		if err != nil {
			logrus.Errorln(err.Error())
			return fmt.Errorf("failed to get issues: %w", err)
		}

		// Fetch group path
		groupPath, err := app.GetGroupPath(groupIDFlag)
		if err != nil {
			logrus.Warnf("Failed to fetch group path: %v", err)
			groupPath = fmt.Sprintf("ID:%d", groupIDFlag)
		}

		// Fetch project paths for all issues
		projectMap, err := app.GetProjectPathsForIssues(issues)
		if err != nil {
			logrus.Warnf("Failed to fetch project paths: %v", err)
			// Fall back to rendering without context
			return renderIssues(issues)
		}

		// Create context and render
		context := render.NewGroupContext(groupPath, projectMap)
		return renderIssuesWithContext(issues, context)
	},
}
