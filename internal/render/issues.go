// Package render provides rendering functionality for GitLab issues
package render

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Maximum title length for display purposes.
const maxTitleLength = 70

// PrintIssues prints the GitLab issues in a formatted table to stdout.
func PrintIssues(issues []*gitlab.Issue, printHeader bool) {
	if printHeader {
		fmt.Printf("%-70s %10s %-25s %-25s\n", "Title", "State", "Created At", "Updated At")
	}
	for idx := range issues {
		// Format issue details with fixed width columns
		title := truncateStr(issues[idx].Title, maxTitleLength)
		state := issues[idx].State
		createdAt := issues[idx].CreatedAt.Format("2006-01-02T15:04:05-0700")
		updatedAt := issues[idx].UpdatedAt.Format("2006-01-02T15:04:05-0700")
		fmt.Printf("%-70s %10s %25s %25s\n", title, state, createdAt, updatedAt)
	}
}

func truncateStr(str string, length int) string {
	if len(str) > length && length > 0 {
		return str[:length]
	}
	return str
}

// PrintTab prints the GitLab issues in a table format using tablewriter.
func PrintTab(issues []*gitlab.Issue) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Title", "State", "CreatedAt", "UpdatedAt"})

	for _, v := range issues {
		t := []string{v.Title, v.State, v.CreatedAt.Format(time.RFC3339), v.UpdatedAt.Format(time.RFC3339)}
		if err := table.Append(t); err != nil {
			fmt.Fprintf(os.Stderr, "Error appending table row: %v\n", err)
		}
	}
	if err := table.Render(); err != nil { // Send output
		fmt.Fprintf(os.Stderr, "Error rendering table: %v\n", err)
	}
}
