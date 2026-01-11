package core

import (
	"errors"
	"fmt"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Static error definitions.
var (
	errMissingIDs     = errors.New("projectID or groupID must be set")
	errConflictingIDs = errors.New("projectID and groupID cannot be set at the same time")
)

// Default pagination value for GitLab API requests.
const defaultPerPage = 100

// GetIssues contains parameters for retrieving issues from GitLab.
type GetIssues struct {
	ProjectID             int64
	GroupID               int64
	State                 string
	FilterCreatedAtAfter  time.Time
	FilterCreatedAtBefore time.Time
	FilterUpdatedAtAfter  time.Time
	FilterUpdatedAtBefore time.Time
	AssigneeUsername      string
}

// GetIssuesOption is a functional option for configuring the GetIssues struct.
type GetIssuesOption func(*GetIssues)

// WithState sets the state filter for issues.
func WithState(state string) GetIssuesOption {
	return func(g *GetIssues) {
		g.State = state
	}
}

// WithFilterCreatedAtAfter filters issues created after the specified time.
func WithFilterCreatedAtAfter(filterCreatedAtAfter time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterCreatedAtAfter = filterCreatedAtAfter
	}
}

// WithFilterCreatedAtBefore filters issues created before the specified time.
func WithFilterCreatedAtBefore(filterCreatedAtBefore time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterCreatedAtBefore = filterCreatedAtBefore
	}
}

// WithFilterUpdatedAtAfter filters issues updated after the specified time.
func WithFilterUpdatedAtAfter(filterUpdatedAtAfter time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterUpdatedAtAfter = filterUpdatedAtAfter
	}
}

// WithFilterUpdatedAtBefore filters issues updated before the specified time.
func WithFilterUpdatedAtBefore(filterUpdatedAtBefore time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterUpdatedAtBefore = filterUpdatedAtBefore
	}
}

// WithProjectID sets the project ID for retrieving issues.
func WithProjectID(projectID int64) GetIssuesOption {
	return func(g *GetIssues) {
		g.ProjectID = projectID
	}
}

// WithGroupID sets the group ID for retrieving issues.
func WithGroupID(groupID int64) GetIssuesOption {
	return func(g *GetIssues) {
		g.GroupID = groupID
	}
}

// WithFilterCreatedAt filters issues by creation date range.
func WithFilterCreatedAt(filterCreatedAtAfter time.Time, filterCreatedAtBefore time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterCreatedAtAfter = filterCreatedAtAfter
		g.FilterCreatedAtBefore = filterCreatedAtBefore
	}
}

// WithFilterUpdatedAt filters issues by update date range.
func WithFilterUpdatedAt(filterUpdatedAtAfter time.Time, filterUpdatedAtBefore time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterUpdatedAtAfter = filterUpdatedAtAfter
		g.FilterUpdatedAtBefore = filterUpdatedAtBefore
	}
}

// WithOpenedIssues filters issues to only show opened issues.
func WithOpenedIssues() GetIssuesOption {
	return func(g *GetIssues) {
		g.State = "opened"
	}
}

// WithClosedIssues filters issues to only show closed issues.
func WithClosedIssues() GetIssuesOption {
	return func(g *GetIssues) {
		g.State = "closed"
	}
}

// WithAssigneeUsername filters issues by assignee username.
func WithAssigneeUsername(assigneeUsername string) GetIssuesOption {
	return func(g *GetIssues) {
		g.AssigneeUsername = assigneeUsername
	}
}

// GetIssues retrieves GitLab issues based on the provided options.
func (a *App) GetIssues(opts ...GetIssuesOption) ([]*gitlab.Issue, error) {
	g := &GetIssues{}
	for _, opt := range opts {
		opt(g)
	}
	if err := g.validate(); err != nil {
		return nil, err
	}
	if g.ProjectID != 0 {
		return a.getIssuesOfProject(g)
	}
	if g.GroupID != 0 {
		return a.getIssuesOfGroup(g)
	}
	return nil, fmt.Errorf("cannot get issues: %w", errMissingIDs)
}

// applyIssueFilters applies common filter settings to issue list options.
func applyIssueFilters(g *GetIssues, listOptions interface{}) {
	// Use type switches to handle both project and group issue options
	switch opts := listOptions.(type) {
	case *gitlab.ListProjectIssuesOptions:
		applyCommonFilters(
			g,
			&opts.State,
			&opts.CreatedAfter,
			&opts.CreatedBefore,
			&opts.UpdatedAfter,
			&opts.UpdatedBefore,
		)
		// Add assignee username filter
		if g.AssigneeUsername != "" {
			opts.AssigneeUsername = &g.AssigneeUsername
		}
	case *gitlab.ListGroupIssuesOptions:
		applyCommonFilters(
			g,
			&opts.State,
			&opts.CreatedAfter,
			&opts.CreatedBefore,
			&opts.UpdatedAfter,
			&opts.UpdatedBefore,
		)
		// Add assignee username filter
		if g.AssigneeUsername != "" {
			opts.AssigneeUsername = &g.AssigneeUsername
		}
	}
}

// applyCommonFilters sets common filter options based on GetIssues fields
// applyCommonFilters sets filter options for different GitLab API option types.
func applyCommonFilters(
	g *GetIssues,
	state **string,
	createdAfter, createdBefore, updatedAfter, updatedBefore **time.Time,
) {
	setStringFilter(state, g.State)
	setTimeFilter(createdAfter, g.FilterCreatedAtAfter)
	setTimeFilter(createdBefore, g.FilterCreatedAtBefore)
	setTimeFilter(updatedAfter, g.FilterUpdatedAtAfter)
	setTimeFilter(updatedBefore, g.FilterUpdatedAtBefore)
}

// setStringFilter safely sets a string filter if the pointer is not nil and the value is not empty.
func setStringFilter(target **string, value string) {
	if target != nil && value != "" {
		*target = &value
	}
}

// setTimeFilter safely sets a time filter if the pointer is not nil and the value is not zero.
func setTimeFilter(target **time.Time, value time.Time) {
	if target != nil && !value.IsZero() {
		*target = &value
	}
}

func (a *App) getIssuesOfProject(g *GetIssues) ([]*gitlab.Issue, error) {
	var allIssues []*gitlab.Issue
	listOptions := gitlab.ListProjectIssuesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: defaultPerPage,
			Page:    1,
		},
	}

	// Apply filters
	applyIssueFilters(g, &listOptions)

	for {
		issues, resp, err := a.gitlabClient.Issues.ListProjectIssues(g.ProjectID, &listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list project issues: %w", err)
		}
		allIssues = append(allIssues, issues...)
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page++
	}
	return allIssues, nil
}

func (a *App) getIssuesOfGroup(g *GetIssues) ([]*gitlab.Issue, error) {
	var allIssues []*gitlab.Issue
	listOptions := gitlab.ListGroupIssuesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: defaultPerPage,
			Page:    1,
		},
	}

	// Apply filters
	applyIssueFilters(g, &listOptions)

	for {
		issues, resp, err := a.gitlabClient.Issues.ListGroupIssues(g.GroupID, &listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list group issues: %w", err)
		}
		allIssues = append(allIssues, issues...)
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page++
	}
	return allIssues, nil
}

func (g *GetIssues) validate() error {
	if g.ProjectID == 0 && g.GroupID == 0 {
		return fmt.Errorf("validation failed: %w", errMissingIDs)
	}
	if g.ProjectID != 0 && g.GroupID != 0 {
		return fmt.Errorf("validation failed: %w", errConflictingIDs)
	}
	return nil
}
