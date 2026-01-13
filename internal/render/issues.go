// Package render provides rendering functionality for GitLab issues
package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Maximum title length for display purposes.
const maxTitleLength = 70

// Maximum title length when displaying project column (reduced to fit both columns).
const maxTitleLengthWithProject = 30

// Renderer defines the interface for rendering GitLab issues.
type Renderer interface {
	Render(issues []*gitlab.Issue, writer io.Writer) error
	RenderWithContext(issues []*gitlab.Issue, context *Context, writer io.Writer) error
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

// RenderWithContext renders issues in plain text format with contextual information.
func (p *PlainRenderer) RenderWithContext(issues []*gitlab.Issue, context *Context, writer io.Writer) error {
	// Write context header if provided
	if context != nil {
		if err := p.writeContextHeader(context, writer); err != nil {
			return err
		}
	}

	// For group queries, add project column
	if context != nil && context.Source == SourceTypeGroup {
		return p.renderWithProjectColumn(issues, context, writer)
	}

	// Fall back to regular rendering for project or no context
	return p.Render(issues, writer)
}

// writeContextHeader writes the context header line.
func (p *PlainRenderer) writeContextHeader(context *Context, writer io.Writer) error {
	switch context.Source {
	case SourceTypeProject:
		if _, err := fmt.Fprintf(writer, "Project: %s\n\n", context.ProjectPath); err != nil {
			return fmt.Errorf("failed to write project header: %w", err)
		}
	case SourceTypeGroup:
		if _, err := fmt.Fprintf(writer, "Group: %s\n\n", context.GroupPath); err != nil {
			return fmt.Errorf("failed to write group header: %w", err)
		}
	}
	return nil
}

// renderWithProjectColumn renders issues with a project column.
func (p *PlainRenderer) renderWithProjectColumn(
	issues []*gitlab.Issue,
	context *Context,
	writer io.Writer,
) error {
	if p.printHeader {
		headerFormat := "%-40s %-30s %10s %-12s %-12s\n"
		if _, err := fmt.Fprintf(writer, headerFormat, "Project", "Title", "State", "Created At", "Updated At"); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	for _, issue := range issues {
		projectPath := context.ProjectMap[issue.ProjectID]
		if projectPath == "" {
			projectPath = fmt.Sprintf("ID:%d", issue.ProjectID)
		}

		title := truncateStr(issue.Title, maxTitleLengthWithProject)
		rowFormat := "%-40s %-30s %10s %12s %12s\n"
		if _, err := fmt.Fprintf(writer, rowFormat,
			projectPath,
			title,
			issue.State,
			issue.CreatedAt.Format("2006-01-02"),
			issue.UpdatedAt.Format("2006-01-02")); err != nil {
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

// RenderWithContext renders issues in table format with contextual information.
func (t *TableRenderer) RenderWithContext(issues []*gitlab.Issue, context *Context, writer io.Writer) error {
	// Write context header if provided
	if context != nil {
		if err := t.writeContextHeader(context, writer); err != nil {
			return err
		}
	}

	table := tablewriter.NewWriter(writer)

	// Adjust headers and data based on context
	if context != nil && context.Source == SourceTypeGroup {
		return t.renderGroupTable(table, issues, context)
	}

	return t.renderRegularTable(table, issues)
}

// renderGroupTable renders the table with project column for group context.
func (t *TableRenderer) renderGroupTable(
	table *tablewriter.Table,
	issues []*gitlab.Issue,
	context *Context,
) error {
	table.Header([]string{"Project", "Title", "State", "CreatedAt", "UpdatedAt"})

	for _, issue := range issues {
		projectPath := context.ProjectMap[issue.ProjectID]
		if projectPath == "" {
			projectPath = fmt.Sprintf("ID:%d", issue.ProjectID)
		}
		row := []string{
			projectPath,
			issue.Title,
			issue.State,
			issue.CreatedAt.Format("2006-01-02"),
			issue.UpdatedAt.Format("2006-01-02"),
		}
		if err := table.Append(row); err != nil {
			return fmt.Errorf("error appending table row: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("error rendering table: %w", err)
	}
	return nil
}

// renderRegularTable renders the table without project column.
func (t *TableRenderer) renderRegularTable(table *tablewriter.Table, issues []*gitlab.Issue) error {
	table.Header([]string{"Title", "State", "CreatedAt", "UpdatedAt"})
	for _, issue := range issues {
		row := []string{
			issue.Title,
			issue.State,
			issue.CreatedAt.Format("2006-01-02"),
			issue.UpdatedAt.Format("2006-01-02"),
		}
		if err := table.Append(row); err != nil {
			return fmt.Errorf("error appending table row: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("error rendering table: %w", err)
	}
	return nil
}

// writeContextHeader writes the context header line for table renderer.
func (t *TableRenderer) writeContextHeader(context *Context, writer io.Writer) error {
	switch context.Source {
	case SourceTypeProject:
		if _, err := fmt.Fprintf(writer, "Project: %s\n\n", context.ProjectPath); err != nil {
			return fmt.Errorf("failed to write project header: %w", err)
		}
	case SourceTypeGroup:
		if _, err := fmt.Fprintf(writer, "Group: %s\n\n", context.GroupPath); err != nil {
			return fmt.Errorf("failed to write group header: %w", err)
		}
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

// RenderWithContext renders issues in markdown format with contextual information.
func (m *MarkdownRenderer) RenderWithContext(issues []*gitlab.Issue, context *Context, writer io.Writer) error {
	// Generate title with context
	title := "# GitLab Issues Report\n\n"
	if context != nil {
		if context.Source == SourceTypeProject {
			title = fmt.Sprintf("# GitLab Issues Report - %s\n\n", context.ProjectPath)
		} else {
			title = fmt.Sprintf("# GitLab Issues Report - Group: %s\n\n", context.GroupPath)
		}
	}

	if len(issues) == 0 {
		if _, err := fmt.Fprintf(writer, "%sNo issues found.\n", title); err != nil {
			return fmt.Errorf("failed to write empty message: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(writer, "%s", title); err != nil {
		return fmt.Errorf("failed to write title: %w", err)
	}

	// For group context, add project column
	if context != nil && context.Source == SourceTypeGroup {
		return m.renderWithProjectColumn(issues, context, writer)
	}

	// Regular rendering without project column
	return m.renderTable(issues, writer)
}

// renderTable renders the markdown table for issues.
func (m *MarkdownRenderer) renderTable(issues []*gitlab.Issue, writer io.Writer) error {
	if _, err := fmt.Fprintf(writer, "| Title | State | Created At | Updated At |\n"); err != nil {
		return fmt.Errorf("failed to write table header: %w", err)
	}
	if _, err := fmt.Fprintf(writer, "|-------|-------|------------|------------|\n"); err != nil {
		return fmt.Errorf("failed to write table separator: %w", err)
	}

	for _, issue := range issues {
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

// renderWithProjectColumn renders the markdown table with a project column.
func (m *MarkdownRenderer) renderWithProjectColumn(
	issues []*gitlab.Issue,
	context *Context,
	writer io.Writer,
) error {
	if _, err := fmt.Fprintf(writer, "| Project | Title | State | Created At | Updated At |\n"); err != nil {
		return fmt.Errorf("failed to write table header: %w", err)
	}
	if _, err := fmt.Fprintf(writer, "|---------|-------|-------|------------|------------|\n"); err != nil {
		return fmt.Errorf("failed to write table separator: %w", err)
	}

	for _, issue := range issues {
		projectPath := context.ProjectMap[issue.ProjectID]
		if projectPath == "" {
			projectPath = fmt.Sprintf("ID:%d", issue.ProjectID)
		}

		title := strings.ReplaceAll(issue.Title, "|", "\\|")
		title = strings.ReplaceAll(title, "\n", " ")
		title = strings.ReplaceAll(title, "\r", " ")

		createdAt := issue.CreatedAt.Format("2006-01-02")
		updatedAt := issue.UpdatedAt.Format("2006-01-02")

		if _, err := fmt.Fprintf(
			writer,
			"| %s | %s | %s | %s | %s |\n",
			projectPath,
			title,
			issue.State,
			createdAt,
			updatedAt,
		); err != nil {
			return fmt.Errorf("failed to write issue row: %w", err)
		}
	}

	return nil
}

func truncateStr(str string, length int) string {
	if len(str) > length && length > 0 {
		return str[:length]
	}
	return str
}
