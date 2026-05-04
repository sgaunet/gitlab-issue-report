package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sgaunet/calcdate/calcdatelib"
	"github.com/sgaunet/gitlab-issue-report/internal/core"
	"github.com/sgaunet/gitlab-issue-report/internal/render"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// commandInit holds the shared initialization result for issue commands.
type commandInit struct {
	beginTime time.Time
	endTime   time.Time
	app       *core.App
}

// initIssueCommand runs the common init pipeline for project and group commands:
// flag reconciliation, logging, environment, timeout, timezone, interval parsing,
// and GitLab client creation.
func initIssueCommand(o *commandOptions, cmd *cobra.Command) (*commandInit, error) {
	if err := reconcileFlags(o); err != nil {
		return nil, err
	}
	initTrace(o.logLevel)
	if err := setupEnvironment(); err != nil {
		return nil, err
	}
	applyTimeoutFromEnv(o, cmd.Flags().Changed("api-timeout"))
	applyTimezoneFromEnv(o, cmd.Flags().Changed("timezone"))
	beginTime, endTime, err := parseInterval(o.interval, o.timezone)
	if err != nil {
		return nil, err
	}
	app, err := core.NewApp(os.Getenv("GITLAB_TOKEN"), os.Getenv("GITLAB_URI"), o.apiTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}
	return &commandInit{beginTime: beginTime, endTime: endTime, app: app}, nil
}

var (
	errGitlabTokenNotSet      = errors.New("GITLAB_TOKEN environment variable is not set")
	errInvalidAPITimeoutValue = errors.New("invalid GITLAB_API_TIMEOUT value")
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

// applyTimeoutFromEnv applies timeout from environment variable if flag not explicitly set.
// This must be called after flag parsing in command execution context.
func applyTimeoutFromEnv(o *commandOptions, flagChanged bool) {
	// Only apply environment variable if flag was not explicitly set
	if !flagChanged {
		if timeoutEnv := os.Getenv("GITLAB_API_TIMEOUT"); timeoutEnv != "" {
			parsed, err := time.ParseDuration(timeoutEnv)
			if err != nil {
				logrus.Warnf("%v: '%s', using default %v", errInvalidAPITimeoutValue, timeoutEnv, defaultAPITimeout)
			} else {
				o.apiTimeout = parsed
				logrus.Debugf("Using API timeout from environment: %v", o.apiTimeout)
			}
		}
	}
}

// applyTimezoneFromEnv applies timezone from environment variable if flag not explicitly set.
// This must be called after flag parsing in command execution context.
func applyTimezoneFromEnv(o *commandOptions, flagChanged bool) {
	if !flagChanged {
		if tzEnv := os.Getenv("GITLAB_TIMEZONE"); tzEnv != "" {
			o.timezone = tzEnv
			logrus.Debugf("Using timezone from environment: %s", o.timezone)
		}
	}
}

// createGitlabClient creates a GitLab client with custom HTTP timeout.
func createGitlabClient(token, uri string, timeout time.Duration) (*gitlab.Client, error) {
	httpClient := &http.Client{
		Timeout: timeout,
	}

	client, err := gitlab.NewClient(
		token,
		gitlab.WithBaseURL(uri),
		gitlab.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	logrus.Debugf("GitLab client created with timeout: %v", timeout)
	return client, nil
}

// getCurrentUsername fetches the username of the currently authenticated user.
func getCurrentUsername(timeout time.Duration) (string, error) {
	gitlabClient, err := createGitlabClient(os.Getenv("GITLAB_TOKEN"), os.Getenv("GITLAB_URI"), timeout)
	if err != nil {
		return "", err
	}

	user, _, err := gitlabClient.Users.CurrentUser()
	if err != nil {
		return "", fmt.Errorf("failed to fetch current user information: %w", err)
	}

	logrus.Debugf("Current user: %s (ID: %d)", user.Username, user.ID)
	return user.Username, nil
}

// parseInterval parses the interval flag and returns the begin and end times.
// The tz parameter specifies the timezone for date calculations (e.g., "America/New_York", "UTC").
// An empty tz string uses the system local timezone.
func parseInterval(interval, tz string) (time.Time, time.Time, error) {
	var beginTime, endTime time.Time
	if interval == "" {
		return time.Time{}, time.Time{}, nil
	}

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
func buildIssueOptions(
	o *commandOptions, projectID, groupID int64, beginTime, endTime time.Time,
) ([]core.GetIssuesOption, error) {
	var options []core.GetIssuesOption

	// Add ID options
	options = addIDOptions(options, projectID, groupID)

	// Add date filter options
	options = addDateFilterOptions(o, options, beginTime, endTime)

	// Add status filter options
	options = addStatusFilterOptions(o, options)

	// Add assignee filter options
	var err error
	options, err = addAssigneeFilterOptions(o, options)
	if err != nil {
		return nil, err
	}

	// Add labels filter options
	options = addLabelsFilterOptions(o, options)

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
	o *commandOptions,
	options []core.GetIssuesOption,
	beginTime, endTime time.Time,
) []core.GetIssuesOption {
	if !beginTime.IsZero() {
		switch {
		case o.createdFilter && !o.updatedFilter:
			options = append(options, core.WithFilterCreatedAt(beginTime, endTime))
		case o.updatedFilter && !o.createdFilter:
			options = append(options, core.WithFilterUpdatedAt(beginTime, endTime))
		case !o.createdFilter && !o.updatedFilter:
			// Default behavior: use updated filter when interval is set but no filter specified
			options = append(options, core.WithFilterUpdatedAt(beginTime, endTime))
		}
		// Both set: invalid, but caught by validation in reconcileFlags
	}
	return options
}

// addStatusFilterOptions adds status filter options based on configuration.
func addStatusFilterOptions(o *commandOptions, options []core.GetIssuesOption) []core.GetIssuesOption {
	switch o.stateFilter {
	case "opened":
		options = append(options, core.WithOpenedIssues())
	case "closed":
		options = append(options, core.WithClosedIssues())
	case "all":
		// No filter, return all issues
	case "":
		// No filter specified, default to all issues
	}
	return options
}

// addAssigneeFilterOptions adds assignee filter options based on mine flag.
func addAssigneeFilterOptions(o *commandOptions, options []core.GetIssuesOption) ([]core.GetIssuesOption, error) {
	if o.mineOption {
		username, err := getCurrentUsername(o.apiTimeout)
		if err != nil {
			return nil, err
		}
		options = append(options, core.WithAssigneeUsername(username))
	}
	return options, nil
}

// addLabelsFilterOptions adds labels filter options if any labels are specified.
func addLabelsFilterOptions(o *commandOptions, options []core.GetIssuesOption) []core.GetIssuesOption {
	labels := sanitizeLabels(o.labelsFilter)
	if len(labels) > 0 {
		options = append(options, core.WithLabels(labels))
	}
	return options
}

// sanitizeLabels trims whitespace and drops empty entries.
func sanitizeLabels(in []string) []string {
	out := make([]string, 0, len(in))
	for _, l := range in {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return out
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

// renderIssues renders the issues based on the format flag.
func renderIssues(issues []*gitlab.Issue, format string) error {
	return renderIssuesWithContext(issues, nil, format)
}

// renderIssuesWithContext renders issues with optional rendering context.
func renderIssuesWithContext(issues []*gitlab.Issue, context *render.Context, format string) error {
	var renderer render.Renderer

	switch format {
	case "markdown":
		renderer = render.NewMarkdownRenderer()
	case "table":
		renderer = render.NewTableRenderer()
	case "plain":
		renderer = render.NewPlainRenderer(true)
	default:
		renderer = render.NewPlainRenderer(true)
	}

	// Use context-aware rendering if context is provided
	if context != nil {
		if err := renderer.RenderWithContext(issues, context, os.Stdout); err != nil {
			return fmt.Errorf("failed to render issues: %w", err)
		}
	} else {
		// Fall back to regular rendering
		if err := renderer.Render(issues, os.Stdout); err != nil {
			return fmt.Errorf("failed to render issues: %w", err)
		}
	}

	return nil
}
