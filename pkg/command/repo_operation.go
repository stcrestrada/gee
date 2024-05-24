package command

import (
	"gee/pkg/types"
	"os/exec"
)

type RepoOperation interface {
	Clone(repoName, remoteUrl, repoPath string, rc *types.RunConfig, onFinish func(onFinish *types.CommandOnFinish))
	Status(repoName, repoPath string, rc *types.RunConfig, onFinish func(onFinish *types.CommandOnFinish))
	Pull(repoName, repoPath string, rc *types.RunConfig, onFinish func(onFinish *types.CommandOnFinish))
	GetRemoteURL(repoName, repoPath string, rc *types.RunConfig, onFinish func(onFinish *types.CommandOnFinish))
}

// GitRepoOperation implements RepoOperation with Git commands
type GitRepoOperation struct{}

func (g *GitRepoOperation) Clone(repoName, remoteUrl, repoPath string, rc *types.RunConfig, onFinish func(onFinish *types.CommandOnFinish)) {
	cmd := exec.Command("git", "-C", repoPath, "clone", remoteUrl)
	runGitCommand(cmd, rc, repoName, onFinish)
}

func (g *GitRepoOperation) Status(repoName, repoPath string, rc *types.RunConfig, onFinish func(onFinish *types.CommandOnFinish)) {
	cmd := exec.Command("git", "-c", "color.status=always", "-C", repoPath, "status")
	runGitCommand(cmd, rc, repoName, onFinish)
}

func (g *GitRepoOperation) Pull(repoName, repoPath string, rc *types.RunConfig, onFinish func(onFinish *types.CommandOnFinish)) {
	cmd := exec.Command("git", "-C", repoPath, "pull")
	runGitCommand(cmd, rc, repoName, onFinish)
}

func (g *GitRepoOperation) GetRemoteURL(repoName, repoPath string, rc *types.RunConfig, onFinish func(onFinish *types.CommandOnFinish)) {
	cmd := exec.Command("git", "-C", repoPath, "config", "--get", "remote.origin.url")
	runGitCommand(cmd, rc, repoName, onFinish)
}

// Helper for executing Git commands and handling results
func runGitCommand(cmd *exec.Cmd, rc *types.RunConfig, repoName string, onFinish func(onFinish *types.CommandOnFinish)) {
	cmd.Stdout = rc.StdOut
	cmd.Stderr = rc.StdErr

	onFinishConfig := &types.CommandOnFinish{
		Repo:      repoName,
		RunConfig: rc,
	}

	err := cmd.Run()

	if err != nil {
		onFinishConfig.Failed = true
		onFinishConfig.Error = err
	}

	onFinish(onFinishConfig)
}
