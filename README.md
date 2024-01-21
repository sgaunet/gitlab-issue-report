[![Go Report Card](https://goreportcard.com/badge/github.com/sgaunet/gitlab-issue-report)](https://goreportcard.com/report/github.com/sgaunet/gitlab-issue-report)

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

