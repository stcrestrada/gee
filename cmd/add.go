package cmd

import (
	"gee/pkg/command"
	"gee/pkg/types"
	"gee/pkg/util"
	"github.com/urfave/cli/v2"
	"os"
)

type AddCommand struct {
	Git       command.GitRepoOperation
	RepoUtils *util.RepoUtils
}

func NewAddCommand() *AddCommand {
	repoOp := command.GitRepoOperation{}
	return &AddCommand{
		Git:       repoOp,
		RepoUtils: util.NewRepoUtils(repoOp),
	}
}

func AddCmd() *cli.Command {
	addCmd := NewAddCommand()
	return &cli.Command{
		Name:  "add",
		Usage: "add repo to gee.toml",
		Action: func(c *cli.Context) error {
			return addCmd.Run(c)
		},
	}
}

func (cmd *AddCommand) Run(c *cli.Context) error {
	cwd, err := cmd.GetWorkingDirectory()
	if err != nil {
		return err
	}

	ctx, err := cmd.LoadConfiguration()
	if err != nil {
		return err
	}

	return cmd.RepoUtils.GeeAdd(ctx, cwd)
}

func (cmd *AddCommand) GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

func (cmd *AddCommand) LoadConfiguration() (*types.GeeContext, error) {
	cwd, err := cmd.GetWorkingDirectory()
	if err != nil {
		return nil, err
	}
	return util.NewConfigHelper().LoadConfig(cwd)
}
