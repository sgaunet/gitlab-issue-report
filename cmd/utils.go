package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/sgaunet/calcdate/calcdatelib"
	"github.com/sgaunet/gitlab-issue-report/internal/core"
	"github.com/sgaunet/gitlab-issue-report/internal/render"
	"github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// setupEnvironment ensures required environment variables are set.
func setupEnvironment() {
	// Check GitLab token
	if len(os.Getenv("GITLAB_TOKEN")) == 0 {
		logrus.Errorf("Set GITLAB_TOKEN environment variable")
		os.Exit(1)
	}
	
	// Set default GitLab URI if not provided
	if len(os.Getenv("GITLAB_URI")) == 0 {
		if err := os.Setenv("GITLAB_URI", "https://gitlab.com"); err != nil {
			logrus.Errorf("Failed to set GITLAB_URI: %v", err)
			os.Exit(1)
		}
	}
}

// parseInterval parses the interval flag and returns the begin and end times.
func parseInterval(interval string) (time.Time, time.Time) {
	var beginTime, endTime time.Time
	if interval == "" {
		return time.Time{}, time.Time{}
	}
	
	tz := ""
	dbegin, err := calcdatelib.NewDate(interval, "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	dbegin.SetBeginDate()
	beginTime = dbegin.Time()
	
	dend, err := calcdatelib.NewDate(interval, "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	dend.SetEndDate()
	endTime = dend.Time()
	
	return beginTime, endTime
}

// buildIssueOptions creates the options for retrieving issues.
func buildIssueOptions(projectID, groupID int, beginTime, endTime time.Time) []core.GetIssuesOption {
	var options []core.GetIssuesOption
	
	// Add ID options
	options = addIDOptions(options, projectID, groupID)
	
	// Add date filter options
	options = addDateFilterOptions(options, beginTime, endTime)
	
	// Add status filter options
	options = addStatusFilterOptions(options)
	
	return options
}

// addIDOptions adds project or group ID options.
func addIDOptions(options []core.GetIssuesOption, projectID, groupID int) []core.GetIssuesOption {
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
func renderIssues(issues []*gitlab.Issue) {
	var renderer render.Renderer
	
	if markdownOutput {
		renderer = render.NewMarkdownRenderer()
	} else {
		renderer = render.NewPlainRenderer(true)
	}
	
	if err := renderer.Render(issues, os.Stdout); err != nil {
		logrus.Errorf("Failed to render issues: %v", err)
		os.Exit(1)
	}
}
