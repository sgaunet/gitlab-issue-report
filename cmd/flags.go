package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// reconcileFlags processes flag values and applies flag priority logic.
func reconcileFlags(cmd *cobra.Command) error {
	// Reconcile logging flags - debug/verbose take precedence over log-level
	if debugFlag {
		logLevel = "debug"
	} else if verboseFlag {
		logLevel = "info"
	}

	return validateFlags()
}

// validateFlags validates flag values and combinations.
func validateFlags() error {
	// Validate state enum
	if stateFilter != "" && stateFilter != "opened" &&
		stateFilter != "closed" && stateFilter != "all" {
		return fmt.Errorf("invalid --state value: %s (must be opened, closed, or all)", stateFilter)
	}

	// Validate format enum
	if formatOutput != "plain" && formatOutput != "table" &&
		formatOutput != "markdown" {
		return fmt.Errorf("invalid --format value: %s (must be plain, table, or markdown)", formatOutput)
	}

	// Validate date filters require interval
	if (createdFilter || updatedFilter) && interval == "" {
		return fmt.Errorf("--created or --updated requires --interval to be set")
	}

	// Validate that both created and updated are not set at the same time
	if createdFilter && updatedFilter {
		return fmt.Errorf("--created and --updated cannot be used together, choose one")
	}

	return nil
}
