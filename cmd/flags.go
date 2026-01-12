package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	errAPITimeoutNonPositive = errors.New("--api-timeout must be positive")
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

	// Validate API timeout
	if apiTimeout <= 0 {
		return fmt.Errorf("%w (got %v)", errAPITimeoutNonPositive, apiTimeout)
	}
	if apiTimeout < 5*time.Second {
		logrus.Warnf("--api-timeout is very short (%v), may cause false timeouts", apiTimeout)
	}
	if apiTimeout > 5*time.Minute {
		logrus.Warnf("--api-timeout is very long (%v), consider using a shorter timeout", apiTimeout)
	}

	return nil
}
