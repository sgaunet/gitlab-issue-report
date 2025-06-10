package core

import (
	"errors"
	"fmt"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GetIssues struct {
	ProjectID             int
	GroupID               int
	State                 string
	FilterCreatedAtAfter  time.Time
	FilterCreatedAtBefore time.Time
	FilterUpdatedAtAfter  time.Time
	FilterUpdatedAtBefore time.Time
}

type GetIssuesOption func(*GetIssues)

func WithState(state string) GetIssuesOption {
	return func(g *GetIssues) {
		g.State = state
	}
}

func WithFilterCreatedAtAfter(filterCreatedAtAfter time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterCreatedAtAfter = filterCreatedAtAfter
	}
}

func WithFilterCreatedAtBefore(filterCreatedAtBefore time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterCreatedAtBefore = filterCreatedAtBefore
	}
}

func WithFilterUpdatedAtAfter(filterUpdatedAtAfter time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterUpdatedAtAfter = filterUpdatedAtAfter
	}
}

func WithFilterUpdatedAtBefore(filterUpdatedAtBefore time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterUpdatedAtBefore = filterUpdatedAtBefore
	}
}

func WithProjectID(projectID int) GetIssuesOption {
	return func(g *GetIssues) {
		g.ProjectID = projectID
	}
}

func WithGroupID(groupID int) GetIssuesOption {
	return func(g *GetIssues) {
		g.GroupID = groupID
	}
}

func WithFilterCreatedAt(filterCreatedAtAfter time.Time, filterCreatedAtBefore time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterCreatedAtAfter = filterCreatedAtAfter
		g.FilterCreatedAtBefore = filterCreatedAtBefore
	}
}

func WithFilterUpdatedAt(filterUpdatedAtAfter time.Time, filterUpdatedAtBefore time.Time) GetIssuesOption {
	return func(g *GetIssues) {
		g.FilterUpdatedAtAfter = filterUpdatedAtAfter
		g.FilterUpdatedAtBefore = filterUpdatedAtBefore
	}
}

func WithOpenedIssues() GetIssuesOption {
	return func(g *GetIssues) {
		g.State = "opened"
	}
}

func WithClosedIssues() GetIssuesOption {
	return func(g *GetIssues) {
		g.State = "closed"
	}
}

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
	return nil, errors.New("projectID or groupID must be set")
}

func (a *App) getIssuesOfProject(g *GetIssues) ([]*gitlab.Issue, error) {
	var allIssues []*gitlab.Issue
	listOptions := gitlab.ListProjectIssuesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}
	if g.State != "" {
		listOptions.State = &g.State
	}
	if !g.FilterCreatedAtAfter.IsZero() {
		listOptions.CreatedAfter = &g.FilterCreatedAtAfter
	}
	if !g.FilterCreatedAtBefore.IsZero() {
		listOptions.CreatedBefore = &g.FilterCreatedAtBefore
	}
	if !g.FilterUpdatedAtAfter.IsZero() {
		listOptions.UpdatedAfter = &g.FilterUpdatedAtAfter
	}
	if !g.FilterUpdatedAtBefore.IsZero() {
		listOptions.UpdatedBefore = &g.FilterUpdatedAtBefore
	}
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
			PerPage: 100,
			Page:    1,
		},
	}
	if g.State != "" {
		listOptions.State = &g.State
	}
	if !g.FilterCreatedAtAfter.IsZero() {
		listOptions.CreatedAfter = &g.FilterCreatedAtAfter
	}
	if !g.FilterCreatedAtBefore.IsZero() {
		listOptions.CreatedBefore = &g.FilterCreatedAtBefore
	}
	if !g.FilterUpdatedAtAfter.IsZero() {
		listOptions.UpdatedAfter = &g.FilterUpdatedAtAfter
	}
	if !g.FilterUpdatedAtBefore.IsZero() {
		listOptions.UpdatedBefore = &g.FilterUpdatedAtBefore
	}

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
		return errors.New("projectID or groupID must be set")
	}
	if g.ProjectID != 0 && g.GroupID != 0 {
		return errors.New("projectID and groupID cannot be set at the same time")
	}
	return nil
}
