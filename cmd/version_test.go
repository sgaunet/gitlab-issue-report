package cmd

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func TestWriteVersionInfo(t *testing.T) {
	t.Run("default output includes all metadata", func(t *testing.T) {
		var buf bytes.Buffer
		writeVersionInfo(&buf, false)
		output := buf.String()

		expectedStrings := []string{
			"gitlab-issue-report version " + version,
			"commit: " + commit,
			"built:  " + buildDate,
			"go:     " + runtime.Version(),
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(output, expected) {
				t.Errorf("default output missing %q\nGot:\n%s", expected, output)
			}
		}

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 4 {
			t.Errorf("default output expected 4 lines, got %d\nGot:\n%s", len(lines), output)
		}
	})

	t.Run("short output is version only", func(t *testing.T) {
		var buf bytes.Buffer
		writeVersionInfo(&buf, true)
		output := buf.String()

		if strings.TrimSpace(output) != version {
			t.Errorf("short output = %q, want %q", strings.TrimSpace(output), version)
		}

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 1 {
			t.Errorf("short output expected 1 line, got %d", len(lines))
		}
	})

	t.Run("development defaults display correctly", func(t *testing.T) {
		// The package-level defaults are "development" and "unknown"
		// which are set when no ldflags are provided (i.e., during tests).
		var buf bytes.Buffer
		writeVersionInfo(&buf, false)
		output := buf.String()

		if !strings.Contains(output, "development") {
			t.Error("expected development version in default output")
		}
		if !strings.Contains(output, "unknown") {
			t.Error("expected unknown commit/buildDate in default output")
		}
	})
}
