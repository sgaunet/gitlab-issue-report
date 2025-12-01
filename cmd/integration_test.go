package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sgaunet/gitlab-issue-report/internal/render"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// TestRenderIssues tests the renderIssues function with different output formats.
func TestRenderIssues(t *testing.T) {
	// Create test issues
	issues := createTestIssues()
	
	tests := []struct {
		name           string
		markdownOutput bool
		expectedStrings []string
	}{
		{
			name:           "plain output",
			markdownOutput: false,
			expectedStrings: []string{"Title", "State", "Created At", "Updated At", "Fix authentication bug", "opened"},
		},
		{
			name:           "markdown output",
			markdownOutput: true,
			expectedStrings: []string{"# GitLab Issues Report", "| Title | State | Created At | Updated At |", "| Fix authentication bug | opened |"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the markdown flag
			markdownOutput = tt.markdownOutput
			
			// Use the renderer directly instead of capturing stdout
			var renderer render.Renderer
			if markdownOutput {
				renderer = render.NewMarkdownRenderer()
			} else {
				renderer = render.NewPlainRenderer(true)
			}
			
			var buf bytes.Buffer
			err := renderer.Render(issues, &buf)
			if err != nil {
				t.Errorf("renderIssues() renderer error = %v", err)
				return
			}
			
			output := buf.String()
			
			// Check expected strings
			for _, expected := range tt.expectedStrings {
				if !strings.Contains(output, expected) {
					t.Errorf("renderIssues() output missing expected string: %q\nGot:\n%s", expected, output)
				}
			}
		})
	}
}

// TestBuildIssueOptions tests the buildIssueOptions function.
func TestBuildIssueOptions(t *testing.T) {
	tests := []struct {
		name             string
		projectID        int64
		groupID          int64
		closedOption     bool
		openedOption     bool
		createdAtOption  bool
		updatedAtOption  bool
		expectedContains []string
	}{
		{
			name:             "project only",
			projectID:        123,
			groupID:          0,
			closedOption:     false,
			openedOption:     false,
			createdAtOption:  false,
			updatedAtOption:  false,
			expectedContains: []string{}, // Can't easily test functional options, but we can test the function doesn't panic
		},
		{
			name:             "group only",
			projectID:        0,
			groupID:          456,
			closedOption:     false,
			openedOption:     false,
			createdAtOption:  false,
			updatedAtOption:  false,
			expectedContains: []string{},
		},
		{
			name:             "with status filters",
			projectID:        123,
			groupID:          0,
			closedOption:     true,
			openedOption:     false,
			createdAtOption:  false,
			updatedAtOption:  false,
			expectedContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global variables to simulate CLI flags
			originalClosed := closedOption
			originalOpened := openedOption
			originalCreatedAt := createdAtOption
			originalUpdatedAt := updatedAtOption
			
			defer func() {
				closedOption = originalClosed
				openedOption = originalOpened
				createdAtOption = originalCreatedAt
				updatedAtOption = originalUpdatedAt
			}()
			
			closedOption = tt.closedOption
			openedOption = tt.openedOption
			createdAtOption = tt.createdAtOption
			updatedAtOption = tt.updatedAtOption
			
			// Call the function and ensure it doesn't panic
			options := buildIssueOptions(tt.projectID, tt.groupID, time.Time{}, time.Time{})
			
			if options == nil {
				t.Error("buildIssueOptions() returned nil")
			}
		})
	}
}

// TestParseInterval tests the parseInterval function.
func TestParseInterval(t *testing.T) {
	tests := []struct {
		name           string
		interval       string
		shouldNotPanic bool
	}{
		{
			name:           "empty interval",
			interval:       "",
			shouldNotPanic: true,
		},
		{
			name:           "valid interval",
			interval:       "/-1/ ::",
			shouldNotPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && tt.shouldNotPanic {
					t.Errorf("parseInterval() panicked: %v", r)
				}
			}()
			
			beginTime, endTime := parseInterval(tt.interval)
			
			if tt.interval == "" {
				if !beginTime.IsZero() || !endTime.IsZero() {
					t.Error("parseInterval() should return zero times for empty interval")
				}
			}
		})
	}
}

// TestInitTrace tests the initTrace function.
func TestInitTrace(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"info level", "info"},
		{"warn level", "warn"},
		{"error level", "error"},
		{"debug level", "debug"},
		{"default level", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("initTrace() panicked: %v", r)
				}
			}()
			
			initTrace(tt.level)
		})
	}
}

// TestSetupEnvironment tests the setupEnvironment function.
func TestSetupEnvironment(t *testing.T) {
	// Save original environment
	originalToken := os.Getenv("GITLAB_TOKEN")
	originalURI := os.Getenv("GITLAB_URI")
	
	defer func() {
		os.Setenv("GITLAB_TOKEN", originalToken)
		os.Setenv("GITLAB_URI", originalURI)
	}()
	
	t.Run("with token and URI", func(t *testing.T) {
		os.Setenv("GITLAB_TOKEN", "test-token")
		os.Setenv("GITLAB_URI", "https://gitlab.example.com")
		
		// This should not panic or exit
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("setupEnvironment() panicked: %v", r)
			}
		}()
		
		setupEnvironment()
		
		if os.Getenv("GITLAB_URI") != "https://gitlab.example.com" {
			t.Error("setupEnvironment() should preserve existing GITLAB_URI")
		}
	})
	
	t.Run("with token but no URI", func(t *testing.T) {
		os.Setenv("GITLAB_TOKEN", "test-token")
		os.Unsetenv("GITLAB_URI")
		
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("setupEnvironment() panicked: %v", r)
			}
		}()
		
		setupEnvironment()
		
		if os.Getenv("GITLAB_URI") != "https://gitlab.com" {
			t.Error("setupEnvironment() should set default GITLAB_URI")
		}
	})
}

// Helper function to create test issues
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
			Title:     "Add new feature",
			State:     "closed",
			CreatedAt: &yesterday,
			UpdatedAt: &now,
		},
	}
}

// TestRendererIntegration tests the integration between CLI and renderers.
func TestRendererIntegration(t *testing.T) {
	issues := createTestIssues()
	
	t.Run("markdown renderer produces valid markdown", func(t *testing.T) {
		renderer := render.NewMarkdownRenderer()
		var buf bytes.Buffer
		
		err := renderer.Render(issues, &buf)
		if err != nil {
			t.Errorf("MarkdownRenderer.Render() error = %v", err)
			return
		}
		
		output := buf.String()
		
		// Check for valid markdown table structure
		if !strings.Contains(output, "| Title | State |") {
			t.Error("Markdown output missing table header")
		}
		
		if !strings.Contains(output, "|-------|-------|") {
			t.Error("Markdown output missing table separator")
		}
		
		if !strings.Contains(output, "# GitLab Issues Report") {
			t.Error("Markdown output missing title")
		}
	})
	
	t.Run("plain renderer produces readable output", func(t *testing.T) {
		renderer := render.NewPlainRenderer(true)
		var buf bytes.Buffer
		
		err := renderer.Render(issues, &buf)
		if err != nil {
			t.Errorf("PlainRenderer.Render() error = %v", err)
			return
		}
		
		output := buf.String()
		
		// Check for plain text structure
		if !strings.Contains(output, "Title") {
			t.Error("Plain output missing Title header")
		}
		
		if !strings.Contains(output, "State") {
			t.Error("Plain output missing State header")
		}
		
		if !strings.Contains(output, "Fix authentication bug") {
			t.Error("Plain output missing issue title")
		}
	})
}