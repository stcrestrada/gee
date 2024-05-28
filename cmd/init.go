package cmd

import (
	"gee/pkg/command"
	"gee/pkg/types"
	"gee/pkg/util"
	"github.com/urfave/cli/v2"
	"os"
)

type InitCommand struct {
	Git       command.GitRepoOperation
	RepoUtils *util.RepoUtils
}

func NewInitCommand() *InitCommand {
	repoOp := command.GitRepoOperation{}
	return &InitCommand{
		Git:       repoOp,
		RepoUtils: util.NewRepoUtils(repoOp),
	}
}

func InitCmd() *cli.Command {
	initCmd := NewInitCommand()
	return &cli.Command{
		Name:  "init",
		Usage: "create gee.toml",
		Action: func(c *cli.Context) error {
			return initCmd.Run(c)
		},
	}
}

func (cmd *InitCommand) Run(c *cli.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		util.Warning("Warning: %s \n", err)
		return err
	}
	util.VerboseLog("initializing gee.toml in %s", cwd)
	err = cmd.RepoUtils.GeeCreate(cwd)
	if err != nil {
		return err
	} else {
		util.Info("Created gee.toml in %s \n", cwd)
	}

	// insert dummy data into gee.toml
	geeCtx := cmd.RepoUtils.NewDummyGeeContext(cwd)
	util.VerboseLog("adding dummy data to gee.toml")
	err = cmd.RepoUtils.InsertConfigIntoGeeToml(geeCtx)
	if err != nil {
		return err
	}
	return err
}

func (cmd *InitCommand) GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

func (cmd *InitCommand) LoadConfiguration() (*types.GeeContext, error) {
	cwd, err := cmd.GetWorkingDirectory()
	if err != nil {
		return nil, err
	}
	util.VerboseLog("loaded gee.toml configuration from %s", cwd)
	return util.NewConfigHelper().LoadConfig(cwd)
}
