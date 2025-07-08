package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// createTestIssues creates test issues for testing purposes.
func createTestIssues() []*gitlab.Issue {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	
	return []*gitlab.Issue{
		{
			ID:        1,
			Title:     "Fix authentication bug",
			State:     "opened",
			CreatedAt: &yesterday,
			UpdatedAt: &now,
		},
		{
			ID:        2,
			Title:     "Add new feature | with pipes",
			State:     "closed",
			CreatedAt: &yesterday,
			UpdatedAt: &now,
		},
		{
			ID:        3,
			Title:     "Update documentation\nwith newlines",
			State:     "opened",
			CreatedAt: &yesterday,
			UpdatedAt: &now,
		},
	}
}

func TestMarkdownRenderer_Render(t *testing.T) {
	tests := []struct {
		name     string
		issues   []*gitlab.Issue
		expected []string
	}{
		{
			name:   "empty issues",
			issues: []*gitlab.Issue{},
			expected: []string{
				"# GitLab Issues Report",
				"No issues found.",
			},
		},
		{
			name:   "single issue",
			issues: createTestIssues()[:1],
			expected: []string{
				"# GitLab Issues Report",
				"| Title | State | Created At | Updated At |",
				"|-------|-------|------------|------------|",
				"| Fix authentication bug | opened |",
			},
		},
		{
			name:   "multiple issues with special characters",
			issues: createTestIssues(),
			expected: []string{
				"# GitLab Issues Report",
				"| Title | State | Created At | Updated At |",
				"|-------|-------|------------|------------|",
				"| Fix authentication bug | opened |",
				"| Add new feature \\| with pipes | closed |",
				"| Update documentation with newlines | opened |",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewMarkdownRenderer()
			var buf bytes.Buffer
			
			err := renderer.Render(tt.issues, &buf)
			if err != nil {
				t.Errorf("MarkdownRenderer.Render() error = %v", err)
				return
			}
			
			output := buf.String()
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("MarkdownRenderer.Render() output missing expected string: %q\nGot:\n%s", expected, output)
				}
			}
		})
	}
}

func TestPlainRenderer_Render(t *testing.T) {
	tests := []struct {
		name        string
		issues      []*gitlab.Issue
		printHeader bool
		expected    []string
	}{
		{
			name:        "empty issues with header",
			issues:      []*gitlab.Issue{},
			printHeader: true,
			expected:    []string{"Title", "State", "Created At", "Updated At"},
		},
		{
			name:        "empty issues without header",
			issues:      []*gitlab.Issue{},
			printHeader: false,
			expected:    []string{},
		},
		{
			name:        "single issue with header",
			issues:      createTestIssues()[:1],
			printHeader: true,
			expected:    []string{"Title", "State", "Created At", "Updated At", "Fix authentication bug", "opened"},
		},
		{
			name:        "multiple issues without header",
			issues:      createTestIssues(),
			printHeader: false,
			expected:    []string{"Fix authentication bug", "Add new feature | with pipes", "Update documentation"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewPlainRenderer(tt.printHeader)
			var buf bytes.Buffer
			
			err := renderer.Render(tt.issues, &buf)
			if err != nil {
				t.Errorf("PlainRenderer.Render() error = %v", err)
				return
			}
			
			output := buf.String()
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("PlainRenderer.Render() output missing expected string: %q\nGot:\n%s", expected, output)
				}
			}
		})
	}
}

func TestTableRenderer_Render(t *testing.T) {
	tests := []struct {
		name     string
		issues   []*gitlab.Issue
		expected []string
	}{
		{
			name:   "empty issues",
			issues: []*gitlab.Issue{},
			expected: []string{
				"TITLE", "STATE", "CREATED AT", "UPDATED AT",
			},
		},
		{
			name:   "single issue",
			issues: createTestIssues()[:1],
			expected: []string{
				"TITLE", "STATE", "CREATED AT", "UPDATED AT",
				"Fix authentication bug", "opened",
			},
		},
		{
			name:   "multiple issues",
			issues: createTestIssues(),
			expected: []string{
				"TITLE", "STATE", "CREATED AT", "UPDATED AT",
				"Fix authentication bug", "opened",
				"Add new feature | with pipes", "closed",
				"Update documentation", "opened",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewTableRenderer()
			var buf bytes.Buffer
			
			err := renderer.Render(tt.issues, &buf)
			if err != nil {
				t.Errorf("TableRenderer.Render() error = %v", err)
				return
			}
			
			output := buf.String()
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("TableRenderer.Render() output missing expected string: %q\nGot:\n%s", expected, output)
				}
			}
		})
	}
}

func TestTruncateStr(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		length   int
		expected string
	}{
		{
			name:     "string shorter than length",
			str:      "short",
			length:   10,
			expected: "short",
		},
		{
			name:     "string longer than length",
			str:      "this is a very long string that should be truncated",
			length:   10,
			expected: "this is a ",
		},
		{
			name:     "string equal to length",
			str:      "exact",
			length:   5,
			expected: "exact",
		},
		{
			name:     "zero length",
			str:      "test",
			length:   0,
			expected: "test",
		},
		{
			name:     "negative length",
			str:      "test",
			length:   -1,
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateStr(tt.str, tt.length)
			if result != tt.expected {
				t.Errorf("truncateStr() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestRendererInterface verifies that all renderers implement the Renderer interface.
func TestRendererInterface(t *testing.T) {
	var _ Renderer = NewMarkdownRenderer()
	var _ Renderer = NewPlainRenderer(true)
	var _ Renderer = NewTableRenderer()
}

// BenchmarkMarkdownRenderer benchmarks the markdown renderer.
func BenchmarkMarkdownRenderer(b *testing.B) {
	issues := createTestIssues()
	renderer := NewMarkdownRenderer()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = renderer.Render(issues, &buf)
	}
}

// BenchmarkPlainRenderer benchmarks the plain renderer.
func BenchmarkPlainRenderer(b *testing.B) {
	issues := createTestIssues()
	renderer := NewPlainRenderer(true)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = renderer.Render(issues, &buf)
	}
}

// BenchmarkTableRenderer benchmarks the table renderer.
func BenchmarkTableRenderer(b *testing.B) {
	issues := createTestIssues()
	renderer := NewTableRenderer()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = renderer.Render(issues, &buf)
	}
}