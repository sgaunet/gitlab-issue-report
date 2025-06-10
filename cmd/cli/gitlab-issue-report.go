package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sgaunet/calcdate/calcdatelib"
	"github.com/sgaunet/gitlab-issue-report/internal/core"
	"github.com/sgaunet/gitlab-issue-report/internal/render"
	"github.com/sirupsen/logrus"
)

var version = "development"

func printVersion() {
	fmt.Println(version)
}

// Config holds command-line configuration.
type Config struct {
	DebugLevel      string
	Interval        string
	ProjectID       int
	GroupID         int
	OpenedOption    bool
	ClosedOption    bool
	CreatedAtOption bool
	UpdatedAtOption bool
	VOption         bool
}

// parseFlags parses command-line flags and returns the parsed values.
func parseFlags() Config {
	var config Config
	flag.StringVar(&config.Interval, "i", "", "interval, ex '/-1/ ::' to describe the interval of last month")
	flag.StringVar(&config.DebugLevel, "d", "error", "Debug level (info,warn,debug)")
	flag.BoolVar(&config.VOption, "v", false, "Get version")
	flag.BoolVar(&config.OpenedOption, "opened", false, "only opened issues")
	flag.BoolVar(&config.ClosedOption, "closed", false, "only closed issues")
	// Use created date filter instead of updated date (default)
	flag.BoolVar(&config.CreatedAtOption, "createdAt", false, "issues filtered with created date")

	flag.IntVar(&config.ProjectID, "p", 0, "Project ID to get issues from")
	flag.IntVar(&config.GroupID, "g", 0, "Group ID to get issues from (not compatible with -p option)")
	flag.Parse()
	return config
}

// validateConfig validates the application configuration and exits if invalid.
func validateConfig(config Config) {
	// Validate debug level
	if config.DebugLevel != "info" && config.DebugLevel != "error" && config.DebugLevel != "debug" {
		logrus.Errorf("debuglevel should be info or error or debug\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	
	// Validate project and group ID combinations
	if config.ProjectID != 0 && config.GroupID != 0 {
		fmt.Fprintln(os.Stderr, "-p and -g option are incompatible")
		flag.PrintDefaults()
		os.Exit(1)
	}
	
	// Validate filter options
	if config.CreatedAtOption && config.UpdatedAtOption {
		logrus.Errorln("createdAt and updatedAt options are incompatible")
		os.Exit(1)
	}
}

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

// findProjectID attempts to determine the project ID if not specified.
func findProjectID(projectID, groupID int) int {
	if projectID != 0 || groupID != 0 {
		return projectID
	}
	
	// Try to find git repository and project
	gitFolder, err := findGitRepository()
	if err != nil {
		logrus.Errorf("Folder .git not found")
		os.Exit(1)
	}
	
	// Get remote origin from git config
	configPath := gitFolder + string(os.PathSeparator) + ".git" + string(os.PathSeparator) + "config"
	remoteOrigin := GetRemoteOrigin(configPath)

	project, err := findProject(remoteOrigin)
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}

	logrus.Infoln("Project found: ", project.SSHURLToRepo)
	logrus.Infoln("Project found: ", project.ID)
	return project.ID
}

// buildIssueOptions creates the options for retrieving issues.
func buildIssueOptions(config Config, projectID, groupID int, beginTime, endTime time.Time) []core.GetIssuesOption {
	var options []core.GetIssuesOption
	
	// Add ID options
	options = addIDOptions(options, projectID, groupID)
	
	// Add date filter options
	options = addDateFilterOptions(options, config, beginTime, endTime)
	
	// Add status filter options
	options = addStatusFilterOptions(options, config)
	
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
	config Config,
	beginTime, endTime time.Time,
) []core.GetIssuesOption {
	if !beginTime.IsZero() {
		if config.CreatedAtOption {
			options = append(options, core.WithFilterCreatedAt(beginTime, endTime))
		}
		if config.UpdatedAtOption {
			options = append(options, core.WithFilterUpdatedAt(beginTime, endTime))
		}
	}
	return options
}

// addStatusFilterOptions adds status filter options based on configuration.
func addStatusFilterOptions(options []core.GetIssuesOption, config Config) []core.GetIssuesOption {
	if config.OpenedOption && !config.ClosedOption {
		options = append(options, core.WithOpenedIssues())
	}
	if config.ClosedOption && !config.OpenedOption {
		options = append(options, core.WithClosedIssues())
	}
	return options
}

func main() {
	// Parse command-line flags
	config := parseFlags()
	
	// Handle version flag
	if config.VOption {
		printVersion()
		os.Exit(0)
	}
	
	// Validate configuration
	validateConfig(config)
	
	// Initialize logging
	initTrace(config.DebugLevel)
	
	// Setup environment
	setupEnvironment()
	
	// Parse interval if provided
	beginTime, endTime := parseInterval(config.Interval)
	
	// Find project ID if not specified
	projectID := findProjectID(config.ProjectID, config.GroupID)
	
	// Create GitLab client
	app, err := core.NewApp(os.Getenv("GITLAB_TOKEN"), os.Getenv("GITLAB_URI"))
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}
	
	// Build issue retrieval options
	options := buildIssueOptions(config, projectID, config.GroupID, beginTime, endTime)
	
	// Get and display issues
	issues, err := app.GetIssues(options...)
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}
	
	render.PrintIssues(issues, true)
	os.Exit(0)
}

func initTrace(debugLevel string) {
	// Log as JSON instead of the default ASCII formatter.
	// logrus.SetFormatter(&logrus.JSONFormatter{})
	// logrus.SetFormatter(&logrus.TextFormatter{
	// 	DisableColors: true,
	// 	FullTimestamp: true,
	// })

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
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
