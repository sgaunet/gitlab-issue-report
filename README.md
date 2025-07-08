[![Go Report Card](https://goreportcard.com/badge/github.com/sgaunet/gitlab-issue-report)](https://goreportcard.com/report/github.com/sgaunet/gitlab-issue-report)
[![GitHub release](https://img.shields.io/github/release/sgaunet/gitlab-issue-report.svg)](https://github.com/sgaunet/gitlab-issue-report/releases/latest)
![GitHub Downloads](https://img.shields.io/github/downloads/sgaunet/gitlab-issue-report/total)
[![GoDoc](https://godoc.org/github.com/sgaunet/gitlab-issue-report?status.svg)](https://godoc.org/github.com/sgaunet/gitlab-issue-report)
[![linter CI Status](https://github.com/sgaunet/gitlab-issue-report/actions/workflows/linter.yml/badge.svg)](https://github.com/sgaunet/gitlab-issue-report/actions/workflows/linter.yml)
[![coverage CI Status](https://github.com/sgaunet/gitlab-issue-report/actions/workflows/coverage.yml/badge.svg)](https://github.com/sgaunet/gitlab-issue-report/actions/workflows/coverage.yml)
[![snapshot CI Status](https://github.com/sgaunet/gitlab-issue-report/actions/workflows/snapshot.yml/badge.svg)](https://github.com/sgaunet/gitlab-issue-report/actions/workflows/snapshot.yml)
[![release CI Status](https://github.com/sgaunet/gitlab-issue-report/actions/workflows/release.yml/badge.svg)](https://github.com/sgaunet/gitlab-issue-report/actions/workflows/release.yml)
[![License](https://img.shields.io/github/license/sgaunet/gitlab-issue-report.svg)](LICENSE)

# gitlab-issue-report

Tool to report issues of a gitlab project/group with multiple output formats (plain text, table, markdown).

# Install 

Copy the binary to /usr/local/bin for example. (or another directory which is in your PATH).

# Usage

```
Usage:
  gitlab-issue-report [command]

Available Commands:
  group       Get issues from a GitLab group
  project     Get issues from a GitLab project

Flags:
  -h, --help   Help for gitlab-issue-report

Project Command Flags:
  -c, --closed            Only closed issues
  -r, --createdAt         Issues filtered with created date
  -d, --d string          Debug level (info,warn,debug) (default "error")
  -i, --i string          Interval, ex '/-1/ ::' to describe the interval of last month
  -p, --id int            Project ID to get issues from
  -m, --markdown          Output in markdown format
  -o, --opened            Only opened issues
  -u, --updatedAt         Issues filtered with updated date

Group Command Flags:
  -c, --closed            Only closed issues
  -r, --createdAt         Issues filtered with created date
  -d, --d string          Debug level (info,warn,debug) (default "error")
  -i, --i string          Interval, ex '/-1/ ::' to describe the interval of last month
  -g, --id int            Group ID to get issues from
  -m, --markdown          Output in markdown format
  -o, --opened            Only opened issues
  -u, --updatedAt         Issues filtered with updated date
```

### Examples

```bash
# Get all issues from a project (default plain text output)
gitlab-issue-report project -p 12345

# Get closed issues from a group created in the last month
gitlab-issue-report group -g 67890 -c -r -i "/-1/ ::"

# Get issues with markdown output for easy sharing
gitlab-issue-report project -p 12345 --markdown

# Get opened issues from a group in markdown format
gitlab-issue-report group -g 67890 -o -m

# Get issues filtered by creation date in markdown format
gitlab-issue-report project -p 12345 --createdAt --markdown -i "/-1/ ::"
```

### Output Formats

The tool supports three output formats:

1. **Plain Text** (default): Simple columnar output
2. **Table**: Formatted table with borders using tablewriter
3. **Markdown**: Markdown table format perfect for documentation and reports

#### Markdown Output Example

```markdown
# GitLab Issues Report

| Title | State | Created At | Updated At |
|-------|-------|------------|------------|
| Fix authentication bug | opened | 2024-01-15 | 2024-01-16 |
| Add new feature | closed | 2024-01-10 | 2024-01-14 |
| Update documentation | opened | 2024-01-12 | 2024-01-13 |
```

## Configuration

2 environement variables can be set :

* GITLAB_TOKEN: used to access to private repositories
* GITLAB_URI: to specify another instance of Gitlab (if not set, GITLAB_URI is set to https://gitlab.com)


# Infos

* [Gitlab Issue API](https://docs.gitlab.com/ee/api/issues.html)
* This project uses the [official GitLab API client library for Go](https://gitlab.com/gitlab-org/api/client-go)

# Development

This project is using :

* golang
* [task for development](https://taskfile.dev/#/)
* [goreleaser](https://goreleaser.com/)
* [pre-commit](https://pre-commit.com/)

There are hooks executed in the precommit stage. Once the project cloned on your disk, please install pre-commit:

```
brew install pre-commit
```

Install tools:

```
task install-prereq
```

And install the hooks:

```
task install-pre-commit
```

If you like to launch manually the pre-commmit hook:

```
task pre-commit
```

### Testing

The project includes comprehensive tests for all functionality:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run specific package tests
go test ./internal/render -v
go test ./cmd -v

# Run linter
task linter
```

Tests follow black box testing principles, focusing on public interfaces and ensuring all output formats work correctly.

## Project Status

🟨 **Maintenance Mode**: This project is used by me to get a summary of what I've done on some projects every month or every sprint. While it's a side project with low priority, it has recently been enhanced with markdown output support.

While we are committed to keeping the project's dependencies up-to-date and secure, please note the following:

- New features are unlikely to be added (though markdown output was recently added)
- Bug fixes will be addressed, but not necessarily promptly
- Security updates will be prioritized

## Issues and Bug Reports

We still encourage you to use our issue tracker for:

- 🐛 Reporting critical bugs
- 🔒 Reporting security vulnerabilities
- 🔍 Asking questions about the project

Please check existing issues before creating a new one to avoid duplicates.

## Contributions

🤝 Limited contributions are still welcome.

While we're not actively developing new features, we appreciate contributions that:

- Fix bugs
- Update dependencies
- Improve documentation
- Enhance performance or security

## Support

As this project is in maintenance mode, support may be limited. We appreciate your understanding and patience.
