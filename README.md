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

Tool report issues of a gitlab project.
**The tool is in beta actually, the command line can change, the options too...**

# Install 

Copy the binary to /usr/local/bin for example. (or another directory which is in your PATH).

# Usage

```
Usage of gitlab-issue-report:
  -closed
        only closed issues
  -createdAt
        issues filtered with created date (updated date by default)
  -d string
        Debug level (info,warn,debug) (default "error")
  -g int
        Group ID to get issues from (not compatible with -p option)
  -i string
        interval, ex /-1/ :: to describe ... (default "/-1/ ::")
  -opened
        only opened issues
  -p int
        Project ID to get issues from
  -v    Get version
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
* docker
* [docker buildx](https://github.com/docker/buildx)
* docker manifest
* [goreleaser](https://goreleaser.com/)
* [venom](https://github.com/ovh/venom) : Tests
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

