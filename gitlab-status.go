package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/src-d/go-git.v4"
)

// our configuration layout
type source struct {
	GitlabURL string `json:"gitlab_url"`
	VerifySSL bool   `json:"verify_ssl"`
	APIKey    string `json:"api_key"`
	Group     string `json:"group"`
	Project   string `json:"project"`
}

type params struct {
	StatusName  string `json:"status_name"`
	BuildStatus string `json:"build_status"`
	Repo        string `json:"repo"`
}

type outRequest struct {
	Source source `json:"source"`
	Params params `json:"params"`
}

type version struct {
	Ref string `json:"ref"`
}

type outResponse struct {
	Version version `json:"version"`
}

func main() {
	switch path.Base(os.Args[0]) {
	case "check":
		check()
	case "in":
		in()
	case "out":
		out()
	default:
		log.Fatalf(`this tool must be called as "check","in" or "out", was: %v`, path.Base(os.Args[0]))
	}
}

func check() {
	// we don't support versions, we just return an empty array
	fmt.Print(`[]`)
}

func in() {
	// in is also not supported, so we return an empty object
	fmt.Print(`{}`)
}

// the only real function
func out() {
	// parse stdin
	req := &outRequest{
		Source: source{
			VerifySSL: true,
		},
		Params: params{
			StatusName: "default",
		},
	}

	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(req); err != nil {
		log.Fatalf("Failed to parse input: %v", err)
	}

	switch req.Params.BuildStatus {
	case "pending", "running", "success", "canceled", "failed":
	default:
		log.Fatalf("invalid build status: %s", req.Params.BuildStatus)
	}
	commitstatus := gitlab.BuildStateValue(req.Params.BuildStatus)

	if req.Params.BuildStatus == "" || req.Params.Repo == "" || req.Source.GitlabURL == "" || req.Source.APIKey == "" ||
		req.Source.Group == "" || req.Source.Project == "" {
		log.Fatal("incomplete configuration")
	}

	repopath := fmt.Sprintf("%s/%s", os.Args[1], req.Params.Repo)
	// now, we can do the real magic
	r, err := git.PlainOpen(repopath)
	if err != nil {
		log.Fatalf("Failed to open git repo: %v", err)
	}

	// get the current commit
	head, err := r.Head()
	if err != nil {
		log.Fatalf("Failed to get git HEAD: %v", err)
	}

	// connect to gitlab
	var httpclnt *http.Client
	if !req.Source.VerifySSL {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpclnt = &http.Client{Transport: tr}
	}

	gitlabapi := gitlab.NewClient(httpclnt, req.Source.APIKey)
	if err := gitlabapi.SetBaseURL(req.Source.GitlabURL); err != nil {
		log.Fatalf("Failed to set gitlab URL: %v", err)
	}

	// update the current commit status
	refname, err := ioutil.ReadFile(fmt.Sprintf("%s/.git/ref"))
	var refstr *string
	if err == nil {
		refstr = new(string)
		*refstr = string(refname)
	}
	jobURL := fmt.Sprintf("%s/teams/%s/pipelines/%s/jobs/%s/builds/%s",
		os.Getenv("ATC_EXTERNAL_URL"), os.Getenv("BUILD_TEAM_NAME"), os.Getenv("BUILD_PIPELINE_NAME"), os.Getenv("BUILD_JOB_NAME"),
		os.Getenv("BUILD_NAME"))
	description := fmt.Sprintf("Concourse build %s", os.Getenv("BUILD_ID"))
	mystatus := &gitlab.SetCommitStatusOptions{
		State:       commitstatus,
		Ref:         refstr,
		Name:        &req.Params.StatusName,
		TargetURL:   &jobURL,
		Description: &description,
	}

	if _, _, err := gitlabapi.Commits.SetCommitStatus(
		fmt.Sprintf("%s/%s", req.Source.Group, req.Source.Project), head.Hash().String(), mystatus); err != nil {
		log.Fatalf("failed to update commit status: %v", err)
	}
	resp := &outResponse{Version: version{Ref: head.Hash().String()}}
	if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
		log.Fatalf("failed to encode output: %v", err)
	}
}
