package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sgaunet/calcdate/calcdatelib"
	gitlabissues "github.com/sgaunet/gitlab-issue-report/gitlabIssues"
	"github.com/sirupsen/logrus"
)

var version string = "development"

func printVersion() {
	fmt.Println(version)
}

func main() {
	var (
		debugLevel string
		interval   string
	)
	var (
		openedOption    bool
		closedOption    bool
		createdAtOption bool
		vOption         bool
	)
	// Parameters treatment (except src + dest)
	flag.StringVar(&interval, "i", "/-1/ ::", "interval, ex /-1/ :: to describe ...")
	flag.StringVar(&debugLevel, "d", "info", "Debug level (info,warn,debug)")
	flag.BoolVar(&vOption, "v", false, "Get version")
	flag.BoolVar(&openedOption, "opened", false, "only opened issues")
	flag.BoolVar(&closedOption, "closed", false, "only closed issues")
	flag.BoolVar(&createdAtOption, "createdAt", false, "issues filtered with created date (updated date by default)")
	flag.Parse()

	if vOption {
		printVersion()
		os.Exit(0)
	}

	if debugLevel != "info" && debugLevel != "warn" && debugLevel != "debug" {
		logrus.Errorf("debuglevel should be info or warn or debug\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	initTrace(debugLevel)

	// tz := os.Getenv("TZ")
	tz := ""
	dBegin, err := calcdatelib.CreateDate(interval, "%YYYY/%MM/%DD %hh:%mm:%ss", tz, true, false)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	dEnd, err := calcdatelib.CreateDate(interval, "%YYYY/%MM/%DD %hh:%mm:%ss", tz, false, true)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// 2022-06-24T08:00:00Z
	gitFolder, err := findGitRepository()
	if err != nil {
		logrus.Errorf("Folder .git not found")
		os.Exit(1)
	}
	remoteOrigin := GetRemoteOrigin(gitFolder + string(os.PathSeparator) + ".git" + string(os.PathSeparator) + "config")
	if len(os.Getenv("GITLAB_TOKEN")) == 0 {
		logrus.Errorf("Set GITLAB_TOKEN environment variable")
		os.Exit(1)
	}
	if len(os.Getenv("GITLAB_URI")) == 0 {
		os.Setenv("GITLAB_URI", "https://gitlab.com")
	}
	project, err := findProject(remoteOrigin)
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}

	logrus.Infoln("Project found: ", project.SshUrlToRepo)
	logrus.Infoln("Project found: ", project.Id)

	// fieldFilterAfter := "updated_after"
	// fieldFilterBefore := "updated_before"
	n := gitlabissues.NewRequestIssues()
	n.SetFilterAfter("updated_after", dBegin)
	n.SetFilterBefore("updated_before", dEnd)

	// state := ""
	if openedOption {
		// state = "opened"
		n.SetOptionOpenedIssues()
	}
	if closedOption {
		// state = "closed"
		n.SetOptionOpenedIssues()
	}
	// var rqt string
	if createdAtOption {
		n.SetFilterAfter("created_after", dBegin)
		n.SetFilterBefore("created_before", dEnd)
	}
	// if state != "" {
	// 	rqt = fmt.Sprintf("issues?state=%s&%s=%s&%s=%s&page=1", state, fieldFilterAfter, dBegin.Format(time.RFC3339), fieldFilterBefore, dEnd.Format(time.RFC3339))
	// } else {
	// 	rqt = fmt.Sprintf("issues?%s=%s&%s=%s&page=1", fieldFilterAfter, dBegin.Format(time.RFC3339), fieldFilterBefore, dEnd.Format(time.RFC3339))
	// }
	// fmt.Println(rqt)
	// _, body, _ := gitlabRequest.Request(rqt)
	// _, body, _ := gitlabRequest.Request("issues?state=opened&updated_after=2022-06-24T08:00:00Z")
	// fmt.Println("%v", string(res))
	// Get project vars

	// var issues []gitlabissues.Issue
	issues, err := n.ExecRequest()
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}

	issues.PrintOneLine(true)
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
