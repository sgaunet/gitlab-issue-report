package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sgaunet/calcdate/calcdatelib"
	"github.com/sgaunet/gitlab-issue-report/internal/core"
	"github.com/sgaunet/gitlab-issue-report/internal/render"
	"github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var (
	errGitlabTokenNotSet = errors.New("GITLAB_TOKEN environment variable is not set")
)

// setupEnvironment ensures required environment variables are set.
func setupEnvironment() error {
	// Check GitLab token
	if len(os.Getenv("GITLAB_TOKEN")) == 0 {
		return errGitlabTokenNotSet
	}

	// Set default GitLab URI if not provided
	if len(os.Getenv("GITLAB_URI")) == 0 {
		if err := os.Setenv("GITLAB_URI", "https://gitlab.com"); err != nil {
			return fmt.Errorf("failed to set GITLAB_URI: %w", err)
		}
	}
	return nil
}

// getCurrentUsername fetches the username of the currently authenticated user.
func getCurrentUsername() (string, error) {
	gitlabClient, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"), gitlab.WithBaseURL(os.Getenv("GITLAB_URI")))
	if err != nil {
		return "", fmt.Errorf("failed to create GitLab client: %w", err)
	}

	user, _, err := gitlabClient.Users.CurrentUser()
	if err != nil {
		return "", fmt.Errorf("failed to fetch current user information: %w", err)
	}

	logrus.Debugf("Current user: %s (ID: %d)", user.Username, user.ID)
	return user.Username, nil
}

// parseInterval parses the interval flag and returns the begin and end times.
func parseInterval(interval string) (time.Time, time.Time, error) {
	var beginTime, endTime time.Time
	if interval == "" {
		return time.Time{}, time.Time{}, nil
	}

	tz := ""
	dbegin, err := calcdatelib.NewDate(interval, "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse begin date: %w", err)
	}
	dbegin.SetBeginDate()
	beginTime = dbegin.Time()

	dend, err := calcdatelib.NewDate(interval, "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse end date: %w", err)
	}
	dend.SetEndDate()
	endTime = dend.Time()

	return beginTime, endTime, nil
}

// buildIssueOptions creates the options for retrieving issues.
func buildIssueOptions(projectID, groupID int64, beginTime, endTime time.Time) ([]core.GetIssuesOption, error) {
	var options []core.GetIssuesOption

	// Add ID options
	options = addIDOptions(options, projectID, groupID)

	// Add date filter options
	options = addDateFilterOptions(options, beginTime, endTime)

	// Add status filter options
	options = addStatusFilterOptions(options)

	// Add assignee filter options
	var err error
	options, err = addAssigneeFilterOptions(options)
	if err != nil {
		return nil, err
	}

	return options, nil
}

// addIDOptions adds project or group ID options.
func addIDOptions(options []core.GetIssuesOption, projectID, groupID int64) []core.GetIssuesOption {
	if projectID != 0 {
		options = append(options, core.WithProjectID(projectID))
	}
	if groupID != 0 {
		options = append(options, core.WithGroupID(groupID))
	}
	return options
}

// addDateFilterOptions adds date filter options based on configuration.
func addDateFilterOptions(
	options []core.GetIssuesOption,
	beginTime, endTime time.Time,
) []core.GetIssuesOption {
	if !beginTime.IsZero() {
		if createdAtOption && !updatedAtOption {
			options = append(options, core.WithFilterCreatedAt(beginTime, endTime))
		} else {
			options = append(options, core.WithFilterUpdatedAt(beginTime, endTime))
		}
	}
	return options
}

// addStatusFilterOptions adds status filter options based on configuration.
func addStatusFilterOptions(options []core.GetIssuesOption) []core.GetIssuesOption {
	if openedOption && !closedOption {
		options = append(options, core.WithOpenedIssues())
	}
	if closedOption && !openedOption {
		options = append(options, core.WithClosedIssues())
	}
	return options
}

// addAssigneeFilterOptions adds assignee filter options based on mine flag.
func addAssigneeFilterOptions(options []core.GetIssuesOption) ([]core.GetIssuesOption, error) {
	if mineOption {
		username, err := getCurrentUsername()
		if err != nil {
			return nil, err
		}
		options = append(options, core.WithAssigneeUsername(username))
	}
	return options, nil
}

// initTrace initializes the logging based on debug level.
func initTrace(debugLevel string) {
	// Output to stdout instead of the default stderr
	logrus.SetOutput(os.Stdout)

	switch debugLevel {
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}
}

// renderIssues renders the issues based on the markdown flag.
func renderIssues(issues []*gitlab.Issue) error {
	var renderer render.Renderer

	if markdownOutput {
		renderer = render.NewMarkdownRenderer()
	} else {
		renderer = render.NewPlainRenderer(true)
	}

	if err := renderer.Render(issues, os.Stdout); err != nil {
		return fmt.Errorf("failed to render issues: %w", err)
	}
	return nil
}
