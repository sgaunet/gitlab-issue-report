package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sgaunet/gitlab-issue-report/internal/core"
	"github.com/sgaunet/gitlab-issue-report/internal/render"
	"github.com/spf13/cobra"
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
			// Use the renderer directly instead of capturing stdout
			var renderer render.Renderer
			switch tt.format {
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
		name      string
		opts      commandOptions
		projectID int64
		groupID   int64
	}{
		{
			name: "project only",
			opts: commandOptions{
				formatOutput: "plain",
				apiTimeout:   defaultAPITimeout,
			},
			projectID: 123,
			groupID:   0,
		},
		{
			name: "group only",
			opts: commandOptions{
				formatOutput: "plain",
				apiTimeout:   defaultAPITimeout,
			},
			projectID: 0,
			groupID:   456,
		},
		{
			name: "with status filters",
			opts: commandOptions{
				stateFilter:  "closed",
				formatOutput: "plain",
				apiTimeout:   defaultAPITimeout,
			},
			projectID: 123,
			groupID:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function and ensure it doesn't panic
			options, err := buildIssueOptions(&tt.opts, tt.projectID, tt.groupID, time.Time{}, time.Time{})

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

			beginTime, endTime, err := parseInterval(tt.interval, "")

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
			var buf bytes.Buffer
			var renderer render.Renderer

			switch tt.format {
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
		name        string
		stateFilter string
	}{
		{
			name:        "state opened",
			stateFilter: "opened",
		},
		{
			name:        "state closed",
			stateFilter: "closed",
		},
		{
			name:        "state all",
			stateFilter: "all",
		},
		{
			name:        "empty state",
			stateFilter: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &commandOptions{
				stateFilter: tt.stateFilter,
				apiTimeout:  defaultAPITimeout,
			}

			options := addStatusFilterOptions(o, []core.GetIssuesOption{})

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
		opts          commandOptions
		expectError   bool
		errorContains string
	}{
		{
			name: "valid state opened",
			opts: commandOptions{
				stateFilter:  "opened",
				formatOutput: "plain",
				apiTimeout:   defaultAPITimeout,
			},
			expectError: false,
		},
		{
			name: "invalid state",
			opts: commandOptions{
				stateFilter:  "invalid",
				formatOutput: "plain",
				apiTimeout:   defaultAPITimeout,
			},
			expectError:   true,
			errorContains: "invalid --state",
		},
		{
			name: "invalid format",
			opts: commandOptions{
				formatOutput: "invalid",
				apiTimeout:   defaultAPITimeout,
			},
			expectError:   true,
			errorContains: "invalid --format",
		},
		{
			name: "created filter without interval",
			opts: commandOptions{
				formatOutput:  "plain",
				createdFilter: true,
				apiTimeout:    defaultAPITimeout,
			},
			expectError:   true,
			errorContains: "requires --interval",
		},
		{
			name: "both created and updated filters",
			opts: commandOptions{
				formatOutput:  "plain",
				createdFilter: true,
				updatedFilter: true,
				interval:      "/-1/ ::",
				apiTimeout:    defaultAPITimeout,
			},
			expectError:   true,
			errorContains: "cannot be used together",
		},
		{
			name: "valid with interval and created",
			opts: commandOptions{
				formatOutput:  "plain",
				createdFilter: true,
				interval:      "/-1/ ::",
				apiTimeout:    defaultAPITimeout,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFlags(&tt.opts)

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
		name    string
		opts    commandOptions
		hasTime bool
	}{
		{
			name:    "created filter only",
			opts:    commandOptions{createdFilter: true},
			hasTime: true,
		},
		{
			name:    "updated filter only",
			opts:    commandOptions{updatedFilter: true},
			hasTime: true,
		},
		{
			name:    "no filters with time",
			opts:    commandOptions{},
			hasTime: true,
		},
		{
			name:    "no filters no time",
			opts:    commandOptions{},
			hasTime: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var beginTime, endTime time.Time
			if tt.hasTime {
				beginTime = yesterday
				endTime = now
			}

			options := addDateFilterOptions(&tt.opts, []core.GetIssuesOption{}, beginTime, endTime)

			// Verify function doesn't panic
			if options == nil {
				t.Error("addDateFilterOptions() returned nil")
			}
		})
	}
}

// TestParseIntervalWithTimezone tests parseInterval with various timezone values.
func TestParseIntervalWithTimezone(t *testing.T) {
	tests := []struct {
		name        string
		interval    string
		tz          string
		expectError bool
	}{
		{
			name:        "valid interval with UTC",
			interval:    "/-1/ ::",
			tz:          "UTC",
			expectError: false,
		},
		{
			name:        "valid interval with America/New_York",
			interval:    "/-1/ ::",
			tz:          "America/New_York",
			expectError: false,
		},
		{
			name:        "valid interval with Europe/Paris",
			interval:    "/-1/ ::",
			tz:          "Europe/Paris",
			expectError: false,
		},
		{
			name:        "valid interval with Local",
			interval:    "/-1/ ::",
			tz:          "Local",
			expectError: false,
		},
		{
			name:        "valid interval with empty timezone",
			interval:    "/-1/ ::",
			tz:          "",
			expectError: false,
		},
		{
			name:        "empty interval with timezone",
			interval:    "",
			tz:          "UTC",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beginTime, endTime, err := parseInterval(tt.interval, tt.tz)

			if tt.expectError {
				if err == nil {
					t.Error("parseInterval() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("parseInterval() unexpected error = %v", err)
				}
			}

			if tt.interval == "" {
				if !beginTime.IsZero() || !endTime.IsZero() {
					t.Error("parseInterval() should return zero times for empty interval")
				}
			}
		})
	}
}

// TestValidateTimezone tests the validateTimezone function.
func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name        string
		tz          string
		expectError bool
	}{
		{
			name:        "empty timezone",
			tz:          "",
			expectError: false,
		},
		{
			name:        "valid UTC",
			tz:          "UTC",
			expectError: false,
		},
		{
			name:        "valid IANA timezone",
			tz:          "America/New_York",
			expectError: false,
		},
		{
			name:        "valid Europe timezone",
			tz:          "Europe/Paris",
			expectError: false,
		},
		{
			name:        "valid Local",
			tz:          "Local",
			expectError: false,
		},
		{
			name:        "invalid timezone",
			tz:          "Invalid/Timezone",
			expectError: true,
		},
		{
			name:        "random string",
			tz:          "not-a-timezone",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &commandOptions{timezone: tt.tz, apiTimeout: defaultAPITimeout}
			err := validateTimezone(o)

			if tt.expectError {
				if err == nil {
					t.Error("validateTimezone() expected error but got none")
				}
				if !strings.Contains(err.Error(), "invalid --timezone") {
					t.Errorf("validateTimezone() error = %v, want error containing 'invalid --timezone'", err)
				}
			} else {
				if err != nil {
					t.Errorf("validateTimezone() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestInitIssueCommand tests the shared init pipeline for issue commands.
func TestInitIssueCommand(t *testing.T) {
	// Helper to create a minimal cobra command with required flags.
	newTestCmd := func() *cobra.Command {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("log-level", "error", "")
		cmd.Flags().Bool("debug", false, "")
		cmd.Flags().Bool("verbose", false, "")
		cmd.Flags().String("state", "", "")
		cmd.Flags().String("format", "plain", "")
		cmd.Flags().Bool("created", false, "")
		cmd.Flags().Bool("updated", false, "")
		cmd.Flags().String("interval", "", "")
		cmd.Flags().Duration("api-timeout", defaultAPITimeout, "")
		cmd.Flags().String("timezone", "", "")
		return cmd
	}

	t.Run("succeeds with valid environment", func(t *testing.T) {
		origToken := os.Getenv("GITLAB_TOKEN")
		origURI := os.Getenv("GITLAB_URI")
		defer func() {
			os.Setenv("GITLAB_TOKEN", origToken)
			os.Setenv("GITLAB_URI", origURI)
		}()

		os.Setenv("GITLAB_TOKEN", "test-token")
		os.Setenv("GITLAB_URI", "https://gitlab.example.com")

		o := &commandOptions{
			logLevel:     "error",
			formatOutput: "plain",
			apiTimeout:   defaultAPITimeout,
		}

		result, err := initIssueCommand(o, newTestCmd())
		if err != nil {
			t.Fatalf("initIssueCommand() unexpected error: %v", err)
		}
		if result.app == nil {
			t.Error("initIssueCommand() result.app is nil")
		}
		if !result.beginTime.IsZero() || !result.endTime.IsZero() {
			t.Error("initIssueCommand() expected zero times for empty interval")
		}
	})

	t.Run("fails without GITLAB_TOKEN", func(t *testing.T) {
		origToken := os.Getenv("GITLAB_TOKEN")
		defer func() {
			os.Setenv("GITLAB_TOKEN", origToken)
		}()

		os.Unsetenv("GITLAB_TOKEN")

		o := &commandOptions{
			logLevel:     "error",
			formatOutput: "plain",
			apiTimeout:   defaultAPITimeout,
		}

		_, err := initIssueCommand(o, newTestCmd())
		if err == nil {
			t.Fatal("initIssueCommand() expected error without GITLAB_TOKEN")
		}
	})

	t.Run("parses interval", func(t *testing.T) {
		origToken := os.Getenv("GITLAB_TOKEN")
		origURI := os.Getenv("GITLAB_URI")
		defer func() {
			os.Setenv("GITLAB_TOKEN", origToken)
			os.Setenv("GITLAB_URI", origURI)
		}()

		os.Setenv("GITLAB_TOKEN", "test-token")
		os.Setenv("GITLAB_URI", "https://gitlab.example.com")

		o := &commandOptions{
			logLevel:     "error",
			formatOutput: "plain",
			interval:     "/-1/ ::",
			apiTimeout:   defaultAPITimeout,
		}

		result, err := initIssueCommand(o, newTestCmd())
		if err != nil {
			t.Fatalf("initIssueCommand() unexpected error: %v", err)
		}
		if result.beginTime.IsZero() || result.endTime.IsZero() {
			t.Error("initIssueCommand() expected non-zero times for valid interval")
		}
	})
}

// TestApplyTimezoneFromEnv tests the applyTimezoneFromEnv function.
func TestApplyTimezoneFromEnv(t *testing.T) {
	t.Run("env applied when flag not set", func(t *testing.T) {
		origEnv := os.Getenv("GITLAB_TIMEZONE")
		defer os.Setenv("GITLAB_TIMEZONE", origEnv)

		o := &commandOptions{}
		os.Setenv("GITLAB_TIMEZONE", "Europe/London")

		applyTimezoneFromEnv(o, false)

		if o.timezone != "Europe/London" {
			t.Errorf("applyTimezoneFromEnv() timezone = %q, want %q", o.timezone, "Europe/London")
		}
	})

	t.Run("flag takes priority over env", func(t *testing.T) {
		origEnv := os.Getenv("GITLAB_TIMEZONE")
		defer os.Setenv("GITLAB_TIMEZONE", origEnv)

		o := &commandOptions{timezone: "America/Chicago"}
		os.Setenv("GITLAB_TIMEZONE", "Europe/London")

		applyTimezoneFromEnv(o, true)

		if o.timezone != "America/Chicago" {
			t.Errorf("applyTimezoneFromEnv() timezone = %q, want %q", o.timezone, "America/Chicago")
		}
	})

	t.Run("no env set leaves timezone unchanged", func(t *testing.T) {
		origEnv := os.Getenv("GITLAB_TIMEZONE")
		defer os.Setenv("GITLAB_TIMEZONE", origEnv)

		o := &commandOptions{}
		os.Unsetenv("GITLAB_TIMEZONE")

		applyTimezoneFromEnv(o, false)

		if o.timezone != "" {
			t.Errorf("applyTimezoneFromEnv() timezone = %q, want empty", o.timezone)
		}
	})
}
