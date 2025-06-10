// Package main implements the gitlab-issue-report command-line tool
package main

import (

	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"gopkg.in/ini.v1"
)

// Define static error variables.
var (
	errTokenNotSet     = errors.New("GITLAB_TOKEN environment variable not set")
	errProjectNotFound = errors.New("project not found")
	errGitNotFound     = errors.New(".git not found")
)



type project struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	SSHURLToRepo  string `json:"sshUrlToRepo"`
	HTTPURLToRepo string `json:"httpUrlToRepo"`
}

func findProject(remoteOrigin string) (project, error) {
	projectName := filepath.Base(remoteOrigin)
	projectName = strings.ReplaceAll(projectName, ".git", "")
	log.Infof("Try to find project %s in %s\n", projectName, os.Getenv("GITLAB_URI"))

	gitlabToken := os.Getenv("GITLAB_TOKEN")
	gitlabURI := os.Getenv("GITLAB_URI")
	if gitlabToken == "" {
		return project{}, fmt.Errorf("gitlab token not available: %w", errTokenNotSet)
	}
	if gitlabURI == "" {
		gitlabURI = "https://gitlab.com"
		log.Warnf("GITLAB_URI not set, defaulting to %s", gitlabURI)
	}

	git, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabURI))
	if err != nil {
		log.Errorf("Failed to create GitLab client: %v", err)
		return project{}, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	searchOpts := &gitlab.SearchOptions{}
	// The scope is implicit in the Projects method call
	foundProjects, _, err := git.Search.Projects(projectName, searchOpts)
	if err != nil {
		log.Errorf("Failed to search for project '%s': %v", projectName, err)
		return project{}, fmt.Errorf("failed to search for project '%s': %w", projectName, err)
	}

	for _, p := range foundProjects {
		log.Debugf("Found project: Name=%s, ID=%d, SSHURL=%s, HTTPURL=%s", p.Name, p.ID, p.SSHURLToRepo, p.HTTPURLToRepo)
		if p.SSHURLToRepo == remoteOrigin {
			return project{
				ID:            p.ID,
				Name:          p.Name,
				SSHURLToRepo:  p.SSHURLToRepo,
				HTTPURLToRepo: p.HTTPURLToRepo,
			}, nil
		}
	}
	return project{}, fmt.Errorf("gitlab project lookup failed: %w", errProjectNotFound)
}

func findGitRepository() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for cwd != "/" {
		log.Debugln(cwd)
		stat, err := os.Stat(cwd + string(os.PathSeparator) + ".git")
		if err == nil {
			if stat.IsDir() {
				return cwd, nil // Found git directory
			}
		}
		cwd = filepath.Dir(cwd)
	}
	return "", fmt.Errorf("git repository not found: %w", errGitNotFound)
}

// GetRemoteOrigin retrieves the remote origin URL from the git configuration file.
func GetRemoteOrigin(gitConfigFile string) string {
	cfg, err := ini.Load(gitConfigFile)
	if err != nil {
		log.Errorf("Fail to read file: %v", err)
		os.Exit(1)
	}

	url := cfg.Section("remote \"origin\"").Key("url").String()
	log.Debugln("url:", url)
	return url
}
