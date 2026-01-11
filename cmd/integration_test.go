package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sgaunet/gitlab-issue-report/internal/core"
	"github.com/sgaunet/gitlab-issue-report/internal/render"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// TestRenderIssues tests the renderIssues function with different output formats.
func TestRenderIssues(t *testing.T) {
	// Create test issues
	issues := createTestIssues()

	tests := []struct {
		name            string
		format          string
		expectedStrings []string
	}{
		{
			name:            "plain output",
			format:          "plain",
			expectedStrings: []string{"Title", "State", "Created At", "Updated At", "Fix authentication bug", "opened"},
		},
		{
			name:            "markdown output",
			format:          "markdown",
			expectedStrings: []string{"# GitLab Issues Report", "| Title | State | Created At | Updated At |", "| Fix authentication bug | opened |"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the format flag
			formatOutput = tt.format

			// Use the renderer directly instead of capturing stdout
			var renderer render.Renderer
			switch formatOutput {
			case "markdown":
				renderer = render.NewMarkdownRenderer()
			case "table":
				renderer = render.NewTableRenderer()
			default:
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
		stateFilter      string
		createdFilter    bool
		updatedFilter    bool
		expectedContains []string
	}{
		{
			name:             "project only",
			projectID:        123,
			groupID:          0,
			stateFilter:      "",
			createdFilter:    false,
			updatedFilter:    false,
			expectedContains: []string{}, // Can't easily test functional options, but we can test the function doesn't panic
		},
		{
			name:             "group only",
			projectID:        0,
			groupID:          456,
			stateFilter:      "",
			createdFilter:    false,
			updatedFilter:    false,
			expectedContains: []string{},
		},
		{
			name:             "with status filters",
			projectID:        123,
			groupID:          0,
			stateFilter:      "closed",
			createdFilter:    false,
			updatedFilter:    false,
			expectedContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global variables to simulate CLI flags
			originalState := stateFilter
			originalCreated := createdFilter
			originalUpdated := updatedFilter

			defer func() {
				stateFilter = originalState
				createdFilter = originalCreated
				updatedFilter = originalUpdated
			}()

			stateFilter = tt.stateFilter
			createdFilter = tt.createdFilter
			updatedFilter = tt.updatedFilter

			// Call the function and ensure it doesn't panic
			options, err := buildIssueOptions(tt.projectID, tt.groupID, time.Time{}, time.Time{})

			if err != nil {
				t.Errorf("buildIssueOptions() error = %v", err)
			}

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

			beginTime, endTime, err := parseInterval(tt.interval)

			if err != nil {
				t.Errorf("parseInterval() error = %v", err)
			}

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

		err := setupEnvironment()
		if err != nil {
			t.Errorf("setupEnvironment() error = %v", err)
		}

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

		err := setupEnvironment()
		if err != nil {
			t.Errorf("setupEnvironment() error = %v", err)
		}

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

// TestNewFormatFlag tests the new format flag with all output formats.
func TestNewFormatFlag(t *testing.T) {
	issues := createTestIssues()

	tests := []struct {
		name            string
		format          string
		expectedStrings []string
	}{
		{
			name:            "plain format",
			format:          "plain",
			expectedStrings: []string{"Title", "State", "Fix authentication bug"},
		},
		{
			name:            "table format",
			format:          "table",
			expectedStrings: []string{"Fix authentication bug", "opened"},
		},
		{
			name:            "markdown format",
			format:          "markdown",
			expectedStrings: []string{"# GitLab Issues Report", "| Title | State |"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore formatOutput
			originalFormat := formatOutput
			defer func() { formatOutput = originalFormat }()

			formatOutput = tt.format

			var buf bytes.Buffer
			var renderer render.Renderer

			switch formatOutput {
			case "markdown":
				renderer = render.NewMarkdownRenderer()
			case "table":
				renderer = render.NewTableRenderer()
			case "plain":
				renderer = render.NewPlainRenderer(true)
			}

			err := renderer.Render(issues, &buf)
			if err != nil {
				t.Errorf("Format %s render error = %v", tt.format, err)
				return
			}

			output := buf.String()
			for _, expected := range tt.expectedStrings {
				if !strings.Contains(output, expected) {
					t.Errorf("Format %s output missing expected string: %q", tt.format, expected)
				}
			}
		})
	}
}

// TestStateFilterFunctionality tests the state filter with new flag.
func TestStateFilterFunctionality(t *testing.T) {
	tests := []struct {
		name         string
		stateFilter  string
		expectOpened bool
		expectClosed bool
	}{
		{
			name:         "state opened",
			stateFilter:  "opened",
			expectOpened: true,
			expectClosed: false,
		},
		{
			name:         "state closed",
			stateFilter:  "closed",
			expectOpened: false,
			expectClosed: true,
		},
		{
			name:         "state all",
			stateFilter:  "all",
			expectOpened: false,
			expectClosed: false,
		},
		{
			name:         "empty state",
			stateFilter:  "",
			expectOpened: false,
			expectClosed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore stateFilter
			originalState := stateFilter
			defer func() { stateFilter = originalState }()

			stateFilter = tt.stateFilter

			options := addStatusFilterOptions([]core.GetIssuesOption{})

			// We can't directly inspect the options, but we can verify the function doesn't panic
			if options == nil {
				t.Error("addStatusFilterOptions() returned nil")
			}
		})
	}
}

// TestValidateFlags tests the validateFlags function.
func TestValidateFlags(t *testing.T) {
	tests := []struct {
		name          string
		setup         func()
		expectError   bool
		errorContains string
	}{
		{
			name: "valid state opened",
			setup: func() {
				stateFilter = "opened"
				formatOutput = "plain"
				createdFilter = false
				updatedFilter = false
				interval = ""
			},
			expectError: false,
		},
		{
			name: "invalid state",
			setup: func() {
				stateFilter = "invalid"
				formatOutput = "plain"
			},
			expectError:   true,
			errorContains: "invalid --state",
		},
		{
			name: "invalid format",
			setup: func() {
				stateFilter = ""
				formatOutput = "invalid"
			},
			expectError:   true,
			errorContains: "invalid --format",
		},
		{
			name: "created filter without interval",
			setup: func() {
				stateFilter = ""
				formatOutput = "plain"
				createdFilter = true
				updatedFilter = false
				interval = ""
			},
			expectError:   true,
			errorContains: "requires --interval",
		},
		{
			name: "both created and updated filters",
			setup: func() {
				stateFilter = ""
				formatOutput = "plain"
				createdFilter = true
				updatedFilter = true
				interval = "/-1/ ::"
			},
			expectError:   true,
			errorContains: "cannot be used together",
		},
		{
			name: "valid with interval and created",
			setup: func() {
				stateFilter = ""
				formatOutput = "plain"
				createdFilter = true
				updatedFilter = false
				interval = "/-1/ ::"
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origState := stateFilter
			origFormat := formatOutput
			origCreated := createdFilter
			origUpdated := updatedFilter
			origInterval := interval

			defer func() {
				stateFilter = origState
				formatOutput = origFormat
				createdFilter = origCreated
				updatedFilter = origUpdated
				interval = origInterval
			}()

			// Setup test state
			tt.setup()

			// Run validation
			err := validateFlags()

			if tt.expectError {
				if err == nil {
					t.Error("validateFlags() expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("validateFlags() error = %v, want error containing %q", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("validateFlags() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestDateFilterOptions tests the updated date filter logic.
func TestDateFilterOptions(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	tests := []struct {
		name          string
		createdFilter bool
		updatedFilter bool
		hasTime       bool
	}{
		{
			name:          "created filter only",
			createdFilter: true,
			updatedFilter: false,
			hasTime:       true,
		},
		{
			name:          "updated filter only",
			createdFilter: false,
			updatedFilter: true,
			hasTime:       true,
		},
		{
			name:          "no filters with time",
			createdFilter: false,
			updatedFilter: false,
			hasTime:       true,
		},
		{
			name:          "no filters no time",
			createdFilter: false,
			updatedFilter: false,
			hasTime:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origCreated := createdFilter
			origUpdated := updatedFilter

			defer func() {
				createdFilter = origCreated
				updatedFilter = origUpdated
			}()

			// Set test values
			createdFilter = tt.createdFilter
			updatedFilter = tt.updatedFilter

			var beginTime, endTime time.Time
			if tt.hasTime {
				beginTime = yesterday
				endTime = now
			}

			options := addDateFilterOptions([]core.GetIssuesOption{}, beginTime, endTime)

			// Verify function doesn't panic
			if options == nil {
				t.Error("addDateFilterOptions() returned nil")
			}
		})
	}
}