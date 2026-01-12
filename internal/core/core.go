// Package core provides the core functionality for interacting with GitLab API.
package core

import (
	"fmt"
	"net/http"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// App represents the application structure for interacting with GitLab API.
type App struct {
	gitlabClient *gitlab.Client
}

// NewApp creates a new application instance with GitLab client.
func NewApp(gitlabToken, gitlabURI string, timeout time.Duration) (*App, error) {
	httpClient := &http.Client{
		Timeout: timeout,
	}

	gitlabClient, err := gitlab.NewClient(
		gitlabToken,
		gitlab.WithBaseURL(gitlabURI),
		gitlab.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}
	return &App{
		gitlabClient: gitlabClient,
	}, nil
}
