// Package render provides rendering functionality for GitLab issues
package render

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Maximum title length for display purposes.
const maxTitleLength = 70

// Renderer defines the interface for rendering GitLab issues.
type Renderer interface {
	Render(issues []*gitlab.Issue, writer io.Writer) error
}

// PlainRenderer renders issues in plain text format.
type PlainRenderer struct {
	printHeader bool
}

// NewPlainRenderer creates a new PlainRenderer.
func NewPlainRenderer(printHeader bool) *PlainRenderer {
	return &PlainRenderer{printHeader: printHeader}
}

// Render renders issues in plain text format.
func (p *PlainRenderer) Render(issues []*gitlab.Issue, writer io.Writer) error {
	if p.printHeader {
		headerFormat := "%-70s %10s %-12s %-12s\n"
		if _, err := fmt.Fprintf(writer, headerFormat, "Title", "State", "Created At", "Updated At"); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}
	for idx := range issues {
		title := truncateStr(issues[idx].Title, maxTitleLength)
		state := issues[idx].State
		createdAt := issues[idx].CreatedAt.Format("2006-01-02")
		updatedAt := issues[idx].UpdatedAt.Format("2006-01-02")
		rowFormat := "%-70s %10s %12s %12s\n"
		if _, err := fmt.Fprintf(writer, rowFormat, title, state, createdAt, updatedAt); err != nil {
			return fmt.Errorf("failed to write issue: %w", err)
		}
	}
	return nil
}

// TableRenderer renders issues in table format.
type TableRenderer struct{}

// NewTableRenderer creates a new TableRenderer.
func NewTableRenderer() *TableRenderer {
	return &TableRenderer{}
}

// Render renders issues in table format.
func (t *TableRenderer) Render(issues []*gitlab.Issue, writer io.Writer) error {
	table := tablewriter.NewWriter(writer)
	table.Header([]string{"Title", "State", "CreatedAt", "UpdatedAt"})

	for _, v := range issues {
		row := []string{v.Title, v.State, v.CreatedAt.Format("2006-01-02"), v.UpdatedAt.Format("2006-01-02")}
		if err := table.Append(row); err != nil {
			return fmt.Errorf("error appending table row: %w", err)
		}
	}
	if err := table.Render(); err != nil {
		return fmt.Errorf("error rendering table: %w", err)
	}
	return nil
}

// MarkdownRenderer renders issues in markdown format.
type MarkdownRenderer struct{}

// NewMarkdownRenderer creates a new MarkdownRenderer.
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{}
}

// Render renders issues in markdown format.
func (m *MarkdownRenderer) Render(issues []*gitlab.Issue, writer io.Writer) error {
	if len(issues) == 0 {
		if _, err := fmt.Fprintf(writer, "# GitLab Issues Report\n\nNo issues found.\n"); err != nil {
			return fmt.Errorf("failed to write empty message: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(writer, "# GitLab Issues Report\n\n"); err != nil {
		return fmt.Errorf("failed to write title: %w", err)
	}
	if _, err := fmt.Fprintf(writer, "| Title | State | Created At | Updated At |\n"); err != nil {
		return fmt.Errorf("failed to write table header: %w", err)
	}
	if _, err := fmt.Fprintf(writer, "|-------|-------|------------|------------|\n"); err != nil {
		return fmt.Errorf("failed to write table separator: %w", err)
	}

	for _, issue := range issues {
		// Escape markdown special characters in title
		title := strings.ReplaceAll(issue.Title, "|", "\\|")
		title = strings.ReplaceAll(title, "\n", " ")
		title = strings.ReplaceAll(title, "\r", " ")
		
		createdAt := issue.CreatedAt.Format("2006-01-02")
		updatedAt := issue.UpdatedAt.Format("2006-01-02")
		
		if _, err := fmt.Fprintf(writer, "| %s | %s | %s | %s |\n", title, issue.State, createdAt, updatedAt); err != nil {
			return fmt.Errorf("failed to write issue row: %w", err)
		}
	}
	
	return nil
}

// PrintIssues prints the GitLab issues in a formatted table to stdout.
// Deprecated: Use Renderer interface instead for better testability.
func PrintIssues(issues []*gitlab.Issue, printHeader bool) {
	renderer := NewPlainRenderer(printHeader)
	_ = renderer.Render(issues, os.Stdout)
}

func truncateStr(str string, length int) string {
	if len(str) > length && length > 0 {
		return str[:length]
	}
	return str
}

// PrintTab prints the GitLab issues in a table format using tablewriter.
// Deprecated: Use Renderer interface instead for better testability.
func PrintTab(issues []*gitlab.Issue) {
	renderer := NewTableRenderer()
	_ = renderer.Render(issues, os.Stdout)
}
