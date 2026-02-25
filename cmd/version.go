package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version   = "development"
	commit    = "unknown"
	buildDate = "unknown"

	shortVersionFlag bool
)

// writeVersionInfo writes version information to the given writer.
// When short is true, only the version string is printed.
func writeVersionInfo(w io.Writer, short bool) error {
	if short {
		if _, err := fmt.Fprintln(w, version); err != nil {
			return fmt.Errorf("failed to write version: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, "gitlab-issue-report version %s\n", version); err != nil {
		return fmt.Errorf("failed to write version info: %w", err)
	}
	if _, err := fmt.Fprintf(w, "  commit: %s\n", commit); err != nil {
		return fmt.Errorf("failed to write version info: %w", err)
	}
	if _, err := fmt.Fprintf(w, "  built:  %s\n", buildDate); err != nil {
		return fmt.Errorf("failed to write version info: %w", err)
	}
	if _, err := fmt.Fprintf(w, "  go:     %s\n", runtime.Version()); err != nil {
		return fmt.Errorf("failed to write version info: %w", err)
	}
	return nil
}

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and build information",
	Long: `Print version and build information for gitlab-issue-report.

By default, displays the version, commit hash, build date, and Go version.
Use --short to print only the version string.

Examples:
  gitlab-issue-report version
  gitlab-issue-report version --short`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return writeVersionInfo(os.Stdout, shortVersionFlag)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&shortVersionFlag, "short", false, "Print only the version string")
}
