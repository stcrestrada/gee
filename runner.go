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

func Clone(repoName string, remoteUrl string, repoPath string, rc *RunConfig, onFinish func(onFinish *CommandOnFinish)) {
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
func Status(repoName string, repoPath string, rc *RunConfig, onFinish func(onFinish *CommandOnFinish)) {
	cmd := exec.Command("git", "-c", "color.status=always", "-C", repoPath, "status")

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

func Pull(repoName string, repoPath string, branch string, rc *RunConfig, onFinish func(onFinish *CommandOnFinish)) {
	var cmd *exec.Cmd
	if strings.HasPrefix(branch, "origin") {
		branchName := RemoveOriginFromBranchName(branch)
		cmd = exec.Command("git", "-C", repoPath, "pull", "origin", fmt.Sprintf("%s", branchName))
	} else {
		cmd = exec.Command("git", "-C", repoPath, "pull", "origin", fmt.Sprintf("%s", branch))
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
func BranchName(repoName string, repoPath string, rc *RunConfig, onFinish func(onFinish *CommandOnFinish)) {
	c := fmt.Sprintf("git -C %s symbolic-ref refs/remotes/origin/HEAD | sed s@^refs/remotes/origin/@@", repoPath)

	cmd := exec.Command("bash", "-C", c)
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
