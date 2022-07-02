
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

