package render

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func PrintIssues(issues []*gitlab.Issue, printHeader bool) {
	if printHeader {
		fmt.Printf("%-70s %10s %-25s %-25s\n", "Title", "State", "Created At", "Updated At")
	}
	for idx := range issues {
		fmt.Printf("%-70s %10s %25s %25s\n", truncateStr(issues[idx].Title, 70), issues[idx].State, issues[idx].CreatedAt.Format("2006-01-02T15:04:05-0700"), issues[idx].UpdatedAt.Format("2006-01-02T15:04:05-0700"))
	}
}

func truncateStr(str string, length int) string {
	if len(str) > length && length > 0 {
		return str[:length]
	}
	return str
}

func PrintTab(issues []*gitlab.Issue) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Title", "State", "CreatedAt", "UpdatedAt"})

	for _, v := range issues {
		t := []string{v.Title, v.State, v.CreatedAt.Format(time.RFC3339), v.UpdatedAt.Format(time.RFC3339)}
		table.Append(t)
	}
	table.Render() // Send output
}
