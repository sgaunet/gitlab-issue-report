// Package main is the entry point for gitlab-issue-report application.
package main

import (
	"os"

	"github.com/sgaunet/gitlab-issue-report/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
