package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sgaunet/gitlab-issue-report/internal/core"
	"github.com/sgaunet/gitlab-issue-report/internal/render"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"gopkg.in/ini.v1"
)

// projectCmd represents the project command.
var (
	errGitRepositoryNotFound   = errors.New("git repository not found")
	errGitlabTokenNotAvailable = errors.New("gitlab token not available")
	errGitlabProjectNotFound   = errors.New("gitlab project not found")
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Get issues of a GitLab project.",
	Long:  `Get issues of a GitLab project by ID or automatically detect from git repository.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Reconcile old and new flags.
		if err := reconcileFlags(cmd); err != nil {
			return err
		}

		// Initialize logging with new log level variable.
		initTrace(logLevel)

		// Setup environment.
		if err := setupEnvironment(); err != nil {
			logrus.Errorln(err.Error())
			return err
		}

		// Apply timeout from environment variable if flag not set.
		applyTimeoutFromEnv(cmd.Flags().Changed("api-timeout"))

		// Parse interval if provided.
		beginTime, endTime, err := parseInterval(interval)
		if err != nil {
			logrus.Errorln(err.Error())
			return err
		}

		// Find project ID if not specified.
		finalProjectID := projectIDFlag
		if finalProjectID == 0 {
			finalProjectID, err = findProjectID()
			if err != nil {
				logrus.Errorln(err.Error())
				return err
			}
		}

		// Create GitLab client.
		app, err := core.NewApp(os.Getenv("GITLAB_TOKEN"), os.Getenv("GITLAB_URI"), apiTimeout)
		if err != nil {
			logrus.Errorln(err.Error())
			return fmt.Errorf("failed to create GitLab client: %w", err)
		}

		// Build issue retrieval options.
		options, err := buildIssueOptions(finalProjectID, 0, beginTime, endTime)
		if err != nil {
			logrus.Errorln(err.Error())
			return err
		}

		// Get and display issues.
		issues, err := app.GetIssues(options...)
		if err != nil {
			logrus.Errorln(err.Error())
			return fmt.Errorf("failed to get issues: %w", err)
		}

		// Fetch project path for context
		projectPath, err := app.GetProjectPath(finalProjectID)
		if err != nil {
			logrus.Warnf("Failed to fetch project path: %v", err)
			// Fall back to rendering without context
			return renderIssues(issues)
		}

		// Create context and render
		context := render.NewProjectContext(projectPath)
		return renderIssuesWithContext(issues, context)
	},
}

// findProjectID attempts to determine the project ID if not specified.
func findProjectID() (int64, error) {
	// Try to find git repository and project.
	gitFolder, err := findGitRepository()
	if err != nil {
		return 0, fmt.Errorf("git repository not found: %w", err)
	}

	// Get remote origin from git config.
	configPath := filepath.Join(gitFolder, ".git", "config")
	remoteOrigin, err := getRemoteOrigin(configPath)
	if err != nil {
		return 0, err
	}

	project, err := findProject(remoteOrigin)
	if err != nil {
		return 0, err
	}

	logrus.Infoln("Project found: ", project.SSHURLToRepo)
	logrus.Infoln("Project found: ", project.ID)
	return project.ID, nil
}

// findGitRepository locates the git repository in the current or parent directories.
func findGitRepository() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("%w", errGitRepositoryNotFound)
	}

	for cwd != "/" {
		logrus.Debugln(cwd)
		stat, err := os.Stat(filepath.Join(cwd, ".git"))
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
func getRemoteOrigin(gitConfigFile string) (string, error) {
	cfg, err := ini.Load(gitConfigFile)
	if err != nil {
		return "", fmt.Errorf("failed to read git config file: %w", err)
	}

	url := cfg.Section("remote \"origin\"").Key("url").String()
	logrus.Debugln("url:", url)
	return url, nil
}

type project struct {
	ID            int64  `json:"id"`
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

	git, err := createGitlabClient(gitlabToken, gitlabURI, apiTimeout)
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
