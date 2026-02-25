package cmd

import (
	"errors"
	"fmt"

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
	Short: "Get issues from a GitLab group",
	Long: `Retrieve and display issues from a GitLab group.

A group ID is required and must be specified with the -g flag.

EXAMPLES:
  # Get all issues from a group
  gitlab-issue-report group -g 678

  # Issues created in the last month
  gitlab-issue-report group -g 678 --created -i "/-30/ ::"

  # Closed issues as markdown
  gitlab-issue-report group -g 678 --state closed --format markdown

  # Only issues assigned to you
  gitlab-issue-report group -g 678 --mine`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		init, err := initIssueCommand(cmd)
		if err != nil {
			return err
		}

		// Check if group ID is provided
		if groupIDFlag == 0 {
			if err := cmd.Help(); err != nil {
				logrus.Warnf("Failed to display help: %v", err)
			}
			return errGroupIDRequired
		}

		// Build issue retrieval options
		options, err := buildIssueOptions(0, groupIDFlag, init.beginTime, init.endTime)
		if err != nil {
			return err
		}

		// Get and display issues
		issues, err := init.app.GetIssues(options...)
		if err != nil {
			return fmt.Errorf("failed to get issues: %w", err)
		}

		// Fetch group path
		groupPath, err := init.app.GetGroupPath(groupIDFlag)
		if err != nil {
			logrus.Warnf("Failed to fetch group path: %v", err)
			groupPath = fmt.Sprintf("ID:%d", groupIDFlag)
		}

		// Fetch project paths for all issues
		projectMap, err := init.app.GetProjectPathsForIssues(issues)
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
