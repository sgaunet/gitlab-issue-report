package gitlabissues

import (
	"encoding/json"
	"fmt"
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
}

func NewRequestIssues() *RequestIssues {
	r := RequestIssues{
		uri: "issues?",
	}
	return &r
}

func (r *RequestIssues) SetOptionClosedIssues() {
	r.state = "closed"
}

func (r *RequestIssues) SetOptionOpenedIssues() {
	r.state = "opened"
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

func (r *RequestIssues) Url() string {
	url := fmt.Sprintf("%s?state=%s", r.uri, r.state)
	if r.fieldFilterAfter != "" {
		url = fmt.Sprintf("%s&%s=%s", url, r.fieldFilterAfter, r.valueFilterAfter.Format(time.RFC3339))
	}
	if r.fieldFilterBefore != "" {
		url = fmt.Sprintf("%s&%s=%s", url, r.fieldFilterBefore, r.valueFilterBefore.Format(time.RFC3339))
	}
	if r.pageNumber != 0 {
		url = fmt.Sprintf("%s&page=%d", url, r.pageNumber)
	}
	return url
	//rqt = fmt.Sprintf("issues?state=%s&%s=%s&%s=%s&page=1", state, fieldFilterAfter, dBegin.Format(time.RFC3339), fieldFilterBefore, dEnd.Format(time.RFC3339))
}

func (r *RequestIssues) ExecRequest() (Issues, error) {
	fmt.Println(r.Url())
	_, _, body, _ := gitlabRequest.Request(r.Url())
	// _, body, _ := gitlabRequest.Request("issues?state=opened&updated_after=2022-06-24T08:00:00Z")
	var issues []Issue
	if err := json.Unmarshal(body, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}
