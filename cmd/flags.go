package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	errAPITimeoutNonPositive  = errors.New("--api-timeout must be positive")
	errInvalidStateValue      = errors.New("invalid --state value")
	errInvalidFormatValue     = errors.New("invalid --format value")
	errIntervalRequired       = errors.New("--created or --updated requires --interval to be set")
	errCreatedUpdatedConflict = errors.New("--created and --updated cannot be used together")
)

// reconcileFlags processes flag values and applies flag priority logic.
func reconcileFlags(_ *cobra.Command) error {
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
	if err := validateStateFlag(); err != nil {
		return err
	}
	if err := validateFormatFlag(); err != nil {
		return err
	}
	if err := validateDateFilters(); err != nil {
		return err
	}
	return validateAPITimeout()
}

// validateStateFlag validates the state filter value.
func validateStateFlag() error {
	if stateFilter != "" && stateFilter != "opened" &&
		stateFilter != "closed" && stateFilter != "all" {
		return fmt.Errorf("%w: %s (must be opened, closed, or all)", errInvalidStateValue, stateFilter)
	}
	return nil
}

// validateFormatFlag validates the format output value.
func validateFormatFlag() error {
	if formatOutput != "plain" && formatOutput != "table" &&
		formatOutput != "markdown" {
		return fmt.Errorf("%w: %s (must be plain, table, or markdown)", errInvalidFormatValue, formatOutput)
	}
	return nil
}

// validateDateFilters validates date filter combinations.
func validateDateFilters() error {
	if (createdFilter || updatedFilter) && interval == "" {
		return errIntervalRequired
	}
	if createdFilter && updatedFilter {
		return errCreatedUpdatedConflict
	}
	return nil
}

// validateAPITimeout validates the API timeout value.
func validateAPITimeout() error {
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
