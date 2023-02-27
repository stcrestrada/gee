package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type RunConfig struct {
	StdOut *bytes.Buffer
	StdErr *bytes.Buffer
}

type CommandOnFinish struct {
	Repo      string
	Path      string
	RunConfig *RunConfig
	Failed    bool
	Error     error
}

func Clone(remoteUrl string, repoPath string, repoName string, rc *RunConfig, onFinish func(onFinish *CommandOnFinish)) {
	cmd := exec.Command("git", "-C", repoPath, "clone", remoteUrl)
	cmd.Stderr = rc.StdErr
	cmd.Stdout = rc.StdOut
	err := cmd.Run()

	onFinish(&CommandOnFinish{
		Repo:      repoName,
		RunConfig: rc,
		Failed:    err != nil,
		Error:     err,
	})

}

func Pull(repoName string, rc *RunConfig, onFinish func(onFinish *CommandOnFinish)) {
	var cmd *exec.Cmd
	branch := "main"
	if strings.HasPrefix(branch, "origin") {
		branchName := RemoveOriginFromBranchName(branch)
		cmd = exec.Command("git", "pull", "origin", fmt.Sprintf("%s", branchName))
	} else {
		cmd = exec.Command("git", "pull", "origin", fmt.Sprintf("%s", branch))
	}
	cmd.Stderr = rc.StdErr
	cmd.Stdout = rc.StdOut
	err := cmd.Run()

	onFinish(&CommandOnFinish{
		Repo:      repoName,
		RunConfig: rc,
		Failed:    err != nil,
		Error:     err,
	})

}
func GetMainBranch(repoName string, rc *RunConfig, onFinish func(onFinish *CommandOnFinish)) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Stderr = rc.StdErr
	cmd.Stdout = rc.StdOut
	err := cmd.Run()

	onFinish(&CommandOnFinish{
		Repo:      repoName,
		RunConfig: rc,
		Failed:    err != nil,
		Error:     err,
	})

}
