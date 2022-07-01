package gitlabissues

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/sgaunet/gitlab-issue-report/gitlabRequest"
)

type RequestIssues struct {
	uri               string
	state             string
	fieldFilterAfter  string
	valueFilterAfter  time.Time
	fieldFilterBefore string
	valueFilterBefore time.Time
	pageNumber        uint
	scope             string
}

func NewRequestIssues() *RequestIssues {
	// list issues of a project : /projects/:id/issues
	// list issues of a group   : /groups/:id/issues
	r := RequestIssues{
		uri:   "",
		scope: "all",
	}
	return &r
}

func (r *RequestIssues) SetProjectId(projectId int) {
	r.uri = fmt.Sprintf("projects/%d/issues?", projectId)
}

func (r *RequestIssues) SetGroupId(groupId int) {
	r.uri = fmt.Sprintf("groups/%d/issues?", groupId)
}

func (r *RequestIssues) SetOptionClosedIssues() {
	r.state = "closed"
}

func (r *RequestIssues) SetOptionOpenedIssues() {
	r.state = "opened"
}

func (r *RequestIssues) SetScopeCurrentUser() {
	r.scope = ""
}

func (r *RequestIssues) SetFilterAfter(field string, d time.Time) {
	r.fieldFilterAfter = field
	r.valueFilterAfter = d
}

func (r *RequestIssues) SetFilterBefore(field string, d time.Time) {
	r.fieldFilterBefore = field
	r.valueFilterBefore = d
}

func (r *RequestIssues) SetPage(pageNumber uint) {
	r.pageNumber = pageNumber
}

func (r *RequestIssues) Url() (url string) {
	url = r.uri
	if r.state != "" {
		url = fmt.Sprintf("%s&state=%s", r.uri, r.state)
	}
	if r.fieldFilterAfter != "" {
		url = fmt.Sprintf("%s&%s=%s", url, r.fieldFilterAfter, r.valueFilterAfter.Format(time.RFC3339))
	}
	if r.fieldFilterBefore != "" {
		url = fmt.Sprintf("%s&%s=%s", url, r.fieldFilterBefore, r.valueFilterBefore.Format(time.RFC3339))
	}
	if r.pageNumber != 0 {
		url = fmt.Sprintf("%s&page=%d", url, r.pageNumber)
	}
	if r.scope != "" {
		url = fmt.Sprintf("%s&scope=%s", url, r.scope)
	}
	return url
	//rqt = fmt.Sprintf("issues?state=%s&%s=%s&%s=%s&page=1", state, fieldFilterAfter, dBegin.Format(time.RFC3339), fieldFilterBefore, dEnd.Format(time.RFC3339))
}

func (r *RequestIssues) GetIssues() (Issues, error) {
	if r.uri == "" {
		return nil, errors.New("no project or group specified")
	}
	var (
		issues    []Issue
		allIssues []Issue
		xNextPage uint
	)

	for {
		r.SetPage(xNextPage)
		resp, body, _ := gitlabRequest.Request(r.Url())
		if err := json.Unmarshal(body, &issues); err != nil {
			return nil, err
		}
		allIssues = append(allIssues, issues...)
		// fmt.Println("****", resp.Header.Get("x-total-pages"))
		nextPage, _ := strconv.Atoi(resp.Header.Get("x-next-page"))
		xNextPage = uint(nextPage)
		if xNextPage == 0 {
			break
		}
	}
	return allIssues, nil
}
