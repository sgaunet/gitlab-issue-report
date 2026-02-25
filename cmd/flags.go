package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	errAPITimeoutNonPositive  = errors.New("--api-timeout must be positive")
	errInvalidStateValue      = errors.New("invalid --state value")
	errInvalidFormatValue     = errors.New("invalid --format value")
	errIntervalRequired       = errors.New("--created or --updated requires --interval to be set")
	errCreatedUpdatedConflict = errors.New("--created and --updated cannot be used together")
	errInvalidTimezoneValue   = errors.New("invalid --timezone value")
)

// reconcileFlags processes flag values and applies flag priority logic.
func reconcileFlags(o *commandOptions) error {
	// Reconcile logging flags - debug/verbose take precedence over log-level
	if o.debugFlag {
		o.logLevel = "debug"
	} else if o.verboseFlag {
		o.logLevel = "info"
	}

	return validateFlags(o)
}

// validateFlags validates flag values and combinations.
func validateFlags(o *commandOptions) error {
	if err := validateStateFlag(o); err != nil {
		return err
	}
	if err := validateFormatFlag(o); err != nil {
		return err
	}
	if err := validateDateFilters(o); err != nil {
		return err
	}
	if err := validateAPITimeout(o); err != nil {
		return err
	}
	return validateTimezone(o)
}

// validateStateFlag validates the state filter value.
func validateStateFlag(o *commandOptions) error {
	if o.stateFilter != "" && o.stateFilter != "opened" &&
		o.stateFilter != "closed" && o.stateFilter != "all" {
		return fmt.Errorf("%w: %s (must be opened, closed, or all)", errInvalidStateValue, o.stateFilter)
	}
	return nil
}

// validateFormatFlag validates the format output value.
func validateFormatFlag(o *commandOptions) error {
	if o.formatOutput != "plain" && o.formatOutput != "table" &&
		o.formatOutput != "markdown" {
		return fmt.Errorf("%w: %s (must be plain, table, or markdown)", errInvalidFormatValue, o.formatOutput)
	}
	return nil
}

// validateDateFilters validates date filter combinations.
func validateDateFilters(o *commandOptions) error {
	if (o.createdFilter || o.updatedFilter) && o.interval == "" {
		return errIntervalRequired
	}
	if o.createdFilter && o.updatedFilter {
		return errCreatedUpdatedConflict
	}
	return nil
}

// validateAPITimeout validates the API timeout value.
func validateAPITimeout(o *commandOptions) error {
	if o.apiTimeout <= 0 {
		return fmt.Errorf("%w (got %v)", errAPITimeoutNonPositive, o.apiTimeout)
	}
	if o.apiTimeout < 5*time.Second {
		logrus.Warnf("--api-timeout is very short (%v), may cause false timeouts", o.apiTimeout)
	}
	if o.apiTimeout > 5*time.Minute {
		logrus.Warnf("--api-timeout is very long (%v), consider using a shorter timeout", o.apiTimeout)
	}
	return nil
}

// validateTimezone validates the timezone value using time.LoadLocation.
func validateTimezone(o *commandOptions) error {
	if o.timezone == "" {
		return nil
	}
	if _, err := time.LoadLocation(o.timezone); err != nil {
		return fmt.Errorf("%w: %s", errInvalidTimezoneValue, o.timezone)
	}
	return nil
}
