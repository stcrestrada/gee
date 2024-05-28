package cmd

import (
	"fmt"
	"gee/pkg/command"
	"gee/pkg/types"
	"gee/pkg/util"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

type RemoveCommand struct {
	Git       command.GitRepoOperation
	RepoUtils *util.RepoUtils
}

func NewRemoveCommand() *RemoveCommand {
	repoOp := command.GitRepoOperation{}
	return &RemoveCommand{
		Git:       repoOp,
		RepoUtils: util.NewRepoUtils(repoOp),
	}
}

func RemoveCmd() *cli.Command {
	removeCmd := NewRemoveCommand()
	return &cli.Command{
		Name:  "remove",
		Usage: "remove repo in gee.toml",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repo",
				Aliases: []string{"r"},
				Usage:   "specify the repository name to remove",
			},
		},
		Action: func(c *cli.Context) error {
			return removeCmd.Run(c)
		},
	}
}

func (cmd *RemoveCommand) Run(c *cli.Context) error {
	var err error
	repoName := c.String("repo")
	if repoName == "" {
		repoName, err = cmd.getRepoNameFromGitDir()
		if err != nil {
			return err
		}
		if repoName == "" {
			return util.NewWarning("please specify the repository name to remove")
		}
	}
	util.VerboseLog("removing repository %s", repoName)

	// Load the configuration
	geeCtx, err := cmd.LoadConfiguration()
	if err != nil {
		return err
	}

	repoPath := ""
	for _, repo := range geeCtx.Repos {
		if repo.Name == repoName {
			repoPath = repo.Path
			break
		}
	}
	if repoPath == "" {
		return util.NewWarning(fmt.Sprintf("repository %s not found in the configuration", repoName))
	}

	// update configuration
	newRepos := []types.Repo{}
	for _, repo := range geeCtx.Repos {
		if repo.Name != repoName {
			newRepos = append(newRepos, repo)
		}
	}

	geeCtx.Repos = newRepos
	configHelper := util.NewConfigHelper()

	err = configHelper.SaveConfig(geeCtx.ConfigFilePath, geeCtx)
	if err != nil {
		return util.NewWarning(err.Error())
	}
	return util.NewInfo(fmt.Sprintf("repository %s successfully remove", repoName))
}

func (cmd *RemoveCommand) GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

func (cmd *RemoveCommand) LoadConfiguration() (*types.GeeContext, error) {
	cwd, err := cmd.GetWorkingDirectory()
	if err != nil {
		return nil, err
	}
	util.VerboseLog("loaded gee.toml configuration from %s", cwd)
	return util.NewConfigHelper().LoadConfig(cwd)
}

// Get repository name from .git directory
func (cmd *RemoveCommand) getRepoNameFromGitDir() (string, error) {
	cwd, err := cmd.GetWorkingDirectory()
	if err != nil {
		return "", util.NewWarning(err.Error())
	}
	gitDir := filepath.Join(cwd, ".git")
	exists, err := cmd.RepoUtils.FileExists(gitDir)
	if err != nil {
		return "", fmt.Errorf("error checking .git directory: %v", err)
	}
	if !exists {
		return "", nil
	}
	return filepath.Base(cwd), nil
}
