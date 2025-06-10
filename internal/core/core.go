package core

import (
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type App struct {
	gitlabClient *gitlab.Client
}

func NewApp(gitlabToken, gitlabURI string) (*App, error) {
	gitlabClient, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabURI))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}
	return &App{
		gitlabClient: gitlabClient,
	}, nil
}
