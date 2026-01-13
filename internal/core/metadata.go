package core

import (
	"fmt"

	"github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// GetProjectPath retrieves the path with namespace for a project.
func (a *App) GetProjectPath(projectID int64) (string, error) {
	project, _, err := a.gitlabClient.Projects.GetProject(int(projectID), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get project %d: %w", projectID, err)
	}
	return project.PathWithNamespace, nil
}

// GetGroupPath retrieves the full path for a group.
func (a *App) GetGroupPath(groupID int64) (string, error) {
	group, _, err := a.gitlabClient.Groups.GetGroup(int(groupID), nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get group %d: %w", groupID, err)
	}
	return group.FullPath, nil
}

// GetProjectPathsForIssues builds a map of projectID -> path for all unique projects in issues.
func (a *App) GetProjectPathsForIssues(issues []*gitlab.Issue) (map[int64]string, error) {
	// Collect unique project IDs
	projectIDs := make(map[int64]bool)
	for _, issue := range issues {
		projectIDs[issue.ProjectID] = true
	}

	// Fetch project paths
	projectPaths := make(map[int64]string)
	for projectID := range projectIDs {
		path, err := a.GetProjectPath(projectID)
		if err != nil {
			// Log warning but continue with other projects
			logrus.Warnf("Failed to fetch path for project %d: %v", projectID, err)
			projectPaths[projectID] = fmt.Sprintf("ID:%d", projectID)
			continue
		}
		projectPaths[projectID] = path
	}

	return projectPaths, nil
}
