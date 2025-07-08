// Package cmd provides commands for gitlab-issue-report.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sgaunet/gitlab-issue-report/internal/core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// projectCmd represents the project command.
var (
	errGitRepositoryNotFound    = errors.New("git repository not found")
	errGitlabTokenNotAvailable = errors.New("gitlab token not available")
	errGitlabProjectNotFound   = errors.New("gitlab project not found")
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Get issues of a GitLab project.",
	Long:  `Get issues of a GitLab project by ID or automatically detect from git repository.`,
	Run: func(_ *cobra.Command, _ []string) {
		// Initialize logging.
		initTrace(debugLevel)

		// Setup environment.
		setupEnvironment()

		// Parse interval if provided.
		beginTime, endTime := parseInterval(interval)

		// Find project ID if not specified.
		finalProjectID := projectID
		if finalProjectID == 0 {
			finalProjectID = findProjectID()
		}

		// Create GitLab client.
		app, err := core.NewApp(os.Getenv("GITLAB_TOKEN"), os.Getenv("GITLAB_URI"))
		if err != nil {
			logrus.Errorln(err.Error())
			os.Exit(1)
		}

		// Build issue retrieval options.
		options := buildIssueOptions(finalProjectID, 0, beginTime, endTime)

		// Get and display issues.
		issues, err := app.GetIssues(options...)
		if err != nil {
			logrus.Errorln(err.Error())
			os.Exit(1)
		}

		renderIssues(issues)
	},
}

// findProjectID attempts to determine the project ID if not specified.
func findProjectID() int {
	// Try to find git repository and project.
	gitFolder, err := findGitRepository()
	if err != nil {
		logrus.Errorf("Folder .git not found")
		os.Exit(1)
	}

	// Get remote origin from git config.
	configPath := gitFolder + string(os.PathSeparator) + ".git" + string(os.PathSeparator) + "config"
	remoteOrigin := getRemoteOrigin(configPath)

	project, err := findProject(remoteOrigin)
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}

	logrus.Infoln("Project found: ", project.SSHURLToRepo)
	logrus.Infoln("Project found: ", project.ID)
	return project.ID
}

// findGitRepository locates the git repository in the current or parent directories.
func findGitRepository() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("%w", errGitRepositoryNotFound)
	}

	for cwd != "/" {
		logrus.Debugln(cwd)
		stat, err := os.Stat(cwd + string(os.PathSeparator) + ".git")
		if err == nil {
			if stat.IsDir() {
				return cwd, nil // Found git directory.
			}
		}
		cwd = filepath.Dir(cwd)
	}
	return "", fmt.Errorf("%w", errGitRepositoryNotFound)
}

// getRemoteOrigin retrieves the remote origin URL from the git configuration file.
func getRemoteOrigin(gitConfigFile string) string {
	cfg, err := ini.Load(gitConfigFile)
	if err != nil {
		logrus.Errorf("Fail to read file: %v", err)
		os.Exit(1)
	}

	url := cfg.Section("remote \"origin\"").Key("url").String()
	logrus.Debugln("url:", url)
	return url
}

type project struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	SSHURLToRepo  string `json:"sshUrlToRepo"`
	HTTPURLToRepo string `json:"httpUrlToRepo"`
}

// findProject searches for a project in GitLab based on the remote origin URL.
func findProject(remoteOrigin string) (project, error) {
	projectName := filepath.Base(remoteOrigin)
	projectName = strings.ReplaceAll(projectName, ".git", "")
	logrus.Infof("Try to find project %s in %s\n", projectName, os.Getenv("GITLAB_URI"))

	gitlabToken := os.Getenv("GITLAB_TOKEN")
	gitlabURI := os.Getenv("GITLAB_URI")
	if gitlabToken == "" {
		return project{}, fmt.Errorf("%w", errGitlabTokenNotAvailable)
	}
	if gitlabURI == "" {
		gitlabURI = "https://gitlab.com"
		logrus.Warnf("GITLAB_URI not set, defaulting to %s", gitlabURI)
	}

	git, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabURI))
	if err != nil {
		logrus.Errorf("Failed to create GitLab client: %v", err)
		return project{}, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	searchOpts := &gitlab.SearchOptions{}
	foundProjects, _, err := git.Search.Projects(projectName, searchOpts)
	if err != nil {
		logrus.Errorf("Failed to search for project '%s': %v", projectName, err)
		return project{}, fmt.Errorf("failed to search for project '%s': %w", projectName, err)
	}

	for _, p := range foundProjects {
		logrus.Debugf("Found project: Name=%s, ID=%d, SSHURL=%s, HTTPURL=%s", p.Name, p.ID, p.SSHURLToRepo, p.HTTPURLToRepo)
		if p.SSHURLToRepo == remoteOrigin {
			return project{
				ID:            p.ID,
				Name:          p.Name,
				SSHURLToRepo:  p.SSHURLToRepo,
				HTTPURLToRepo: p.HTTPURLToRepo,
			}, nil
		}
	}
	return project{}, fmt.Errorf("%w", errGitlabProjectNotFound)
}
