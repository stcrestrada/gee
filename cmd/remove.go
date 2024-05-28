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
	util.VerboseLog("attempting to remove %s from gee.toml", repoName)

	// Load the configuration
	geeCtx, err := cmd.LoadConfiguration()
	if err != nil {
		return err
	}

	var index *int
	for i, repo := range geeCtx.Repos {
		if repo.Name == repoName {
			index = &i
			break
		}
	}
	if index == nil {
		return util.NewWarning(fmt.Sprintf("%s not found in gee.toml", repoName))
	}

	repoLength := len(geeCtx.Repos)
	// update repos list in configuration
	geeCtx.Repos[repoLength-1], geeCtx.Repos[*index] = geeCtx.Repos[*index], geeCtx.Repos[repoLength-1]
	geeCtx.Repos = geeCtx.Repos[:repoLength-1]

	configHelper := util.NewConfigHelper()

	err = configHelper.SaveConfig(geeCtx.ConfigFilePath, geeCtx)
	if err != nil {
		return util.NewWarning(err.Error())
	}
	return util.NewInfo(fmt.Sprintf("successfully removed %s from gee.toml", repoName))
}

func (cmd *RemoveCommand) GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

func (cmd *RemoveCommand) LoadConfiguration() (*types.GeeContext, error) {
	cwd, err := cmd.GetWorkingDirectory()
	if err != nil {
		return nil, err
	}
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
