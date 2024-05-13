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

func Pull(repoName string, repoPath string, rc *RunConfig, onFinish func(onFinish *CommandOnFinish)) {
	cmd := exec.Command("git", "-C", repoPath, "pull")
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

func HandlePullFinish(repo *Repo, onFinish *CommandOnFinish, state *SpinnerState) {
	switch {
	case onFinish.Failed && strings.Contains(onFinish.RunConfig.StdErr.String(), "No such file or directory") && repo.Remote != "":
		state.Msg = fmt.Sprintf("Cloning instead...")
		Clone(repo.Name, repo.Remote, repo.Path, onFinish.RunConfig, HandleCloneFinish(repo, state))
	case onFinish.Failed:
		state.State = StateError
		state.Msg = fmt.Sprintf("Failed to pull %s", repo.Name)
	default:
		Clone(repo.Name, repo.Remote, repo.Path, onFinish.RunConfig, HandleCloneFinish(repo, state))
	}
}

func HandleCloneFinish(repo *Repo, state *SpinnerState) func(onFinish *CommandOnFinish) {
	return func(onFinish *CommandOnFinish) {
		switch {
		case onFinish.Failed && strings.Contains(onFinish.RunConfig.StdErr.String(), "already exists"):
			state.State = StateSuccess
			state.Msg = fmt.Sprintf("Already cloned %s", repo.Name)
		case onFinish.Failed:
			state.State = StateError
			state.Msg = fmt.Sprintf("Failed to clone %s", repo.Name)
		default:
			state.State = StateSuccess
			state.Msg = fmt.Sprintf("Finished cloning %s", repo.Name)
		}
	}
}

func GetRemoteURL(repoName string, repoPath string, rc *RunConfig, onFinish func(onFinish *CommandOnFinish)) {
	cmd := exec.Command("git", "-C", repoPath, "config", "--get", "remote.origin.url")
	cmd.Stderr = rc.StdErr
	cmd.Stdout = rc.StdOut
	err := cmd.Run()

	onFinish(&CommandOnFinish{
		Repo:      repoName,
		Path:      repoPath,
		RunConfig: rc,
		Failed:    err != nil,
		Error:     err,
	})
}
