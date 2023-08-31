package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sgaunet/calcdate/calcdatelib"
	gitlabissues "github.com/sgaunet/gitlab-issue-report/pkg/gitlabIssues"
	"github.com/sirupsen/logrus"
)

var version string = "development"

func printVersion() {
	fmt.Println(version)
}

func main() {
	var (
		debugLevel      string
		interval        string
		projectId       int
		groupId         int
		openedOption    bool
		closedOption    bool
		createdAtOption bool
		vOption         bool
		dBegin          time.Time
		dEnd            time.Time
		err             error
	)
	// Parameters treatment (except src + dest)
	flag.StringVar(&interval, "i", "", "interval, ex '/-1/ ::' to describe the interval of last month")
	flag.StringVar(&debugLevel, "d", "error", "Debug level (info,warn,debug)")
	flag.BoolVar(&vOption, "v", false, "Get version")
	flag.BoolVar(&openedOption, "opened", false, "only opened issues")
	flag.BoolVar(&closedOption, "closed", false, "only closed issues")
	flag.BoolVar(&createdAtOption, "createdAt", false, "issues filtered with created date (updated date by default)")
	flag.IntVar(&projectId, "p", 0, "Project ID to get issues from")
	flag.IntVar(&groupId, "g", 0, "Group ID to get issues from (not compatible with -p option)")
	flag.Parse()

	if vOption {
		printVersion()
		os.Exit(0)
	}

	if debugLevel != "info" && debugLevel != "error" && debugLevel != "debug" {
		logrus.Errorf("debuglevel should be info or error or debug\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if projectId != 0 && groupId != 0 {
		fmt.Fprintln(os.Stderr, "-p and -g option are incomptabile")
		flag.PrintDefaults()
		os.Exit(1)
	}
	initTrace(debugLevel)
	if len(os.Getenv("GITLAB_TOKEN")) == 0 {
		logrus.Errorf("Set GITLAB_TOKEN environment variable")
		os.Exit(1)
	}
	if len(os.Getenv("GITLAB_URI")) == 0 {
		os.Setenv("GITLAB_URI", "https://gitlab.com")
	}

	// if option -i , calculdate
	if interval != "" {
		tz := ""
		dbegin, err := calcdatelib.NewDate(interval, "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		dbegin.SetBeginDate()
		dBegin = dbegin.Time()
		dend, err := calcdatelib.NewDate(interval, "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		dend.SetEndDate()
		dEnd = dend.Time()
	}
	if groupId == 0 && projectId == 0 {
		// Try to find git repository and project
		gitFolder, err := findGitRepository()
		if err != nil {
			logrus.Errorf("Folder .git not found")
			os.Exit(1)
		}
		remoteOrigin := GetRemoteOrigin(gitFolder + string(os.PathSeparator) + ".git" + string(os.PathSeparator) + "config")

		project, err := findProject(remoteOrigin)
		if err != nil {
			logrus.Errorln(err.Error())
			os.Exit(1)
		}

		logrus.Infoln("Project found: ", project.SshUrlToRepo)
		logrus.Infoln("Project found: ", project.Id)
		projectId = project.Id
	}

	// fieldFilterAfter := "updated_after"
	// fieldFilterBefore := "updated_before"
	n := gitlabissues.NewRequestIssues()
	if interval != "" {
		// by default, filter date is on updated time
		n.SetFilterAfter("updated_after", dBegin)
		n.SetFilterBefore("updated_before", dEnd)
		if createdAtOption {
			n.SetFilterAfter("created_after", dBegin)
			n.SetFilterBefore("created_before", dEnd)
		}
	}

	if openedOption {
		n.SetOptionOpenedIssues()
	}
	if closedOption {
		n.SetOptionClosedIssues()
	}

	if projectId != 0 {
		n.SetProjectId(projectId)
	}
	if groupId != 0 {
		n.SetGroupId(groupId)
	}
	issues, err := n.GetIssues()
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}
	issues.PrintIssues(true)
}

func initTrace(debugLevel string) {
	// Log as JSON instead of the default ASCII formatter.
	//logrus.SetFormatter(&logrus.JSONFormatter{})
	// logrus.SetFormatter(&logrus.TextFormatter{
	// 	DisableColors: true,
	// 	FullTimestamp: true,
	// })

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	switch debugLevel {
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}
}
