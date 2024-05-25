package cmd

import (
	"bytes"
	"fmt"
	"gee/pkg/command"
	"gee/pkg/types"
	"gee/pkg/ui"
	"gee/pkg/util"
	"github.com/pborman/indent"
	"github.com/stcrestrada/gogo"
	"github.com/urfave/cli/v2"
	"os"
)

type StatusCommand struct {
	Git       command.GitRepoOperation
	RepoUtils *util.RepoUtils
}

func NewStatusCommand() *StatusCommand {
	repoOp := command.GitRepoOperation{}
	return &StatusCommand{
		Git:       repoOp,
		RepoUtils: util.NewRepoUtils(repoOp),
	}
}

func StatusCmd() *cli.Command {
	statusCmd := NewStatusCommand()
	return &cli.Command{
		Name:  "status",
		Usage: "Git status of all repos",
		Action: func(c *cli.Context) error {
			return statusCmd.Run(c)
		},
	}
}

func (cmd *StatusCommand) Run(c *cli.Context) error {
	ctx, err := cmd.LoadConfiguration()
	if err != nil {
		util.Warning("Warning: %s", err)
		return err
	}

	repos := ctx.Repos
	states := make([]*ui.SpinnerState, len(repos))
	commandOnFinish := make([]*types.CommandOnFinish, len(repos))

	for i, repo := range repos {
		states[i] = &ui.SpinnerState{
			State: ui.StateLoading,
			Msg:   fmt.Sprintf("Retrieve status for %s", repo.Name),
		}
	}

	finishPrint := ui.PrintSpinnerStates(os.Stdout, states)

	concurrency := len(repos)
	pool := gogo.NewPool(concurrency, len(repos), func(i int) func() (interface{}, error) {
		repo := repos[i]
		return func() (interface{}, error) {
			fullPath := cmd.RepoUtils.FullPathWithRepo(repo.Path, repo.Name)

			rc := &types.RunConfig{
				StdErr: &bytes.Buffer{},
				StdOut: &bytes.Buffer{},
			}

			cmd.Git.Status(repo.Name, fullPath, rc, func(onFinish *types.CommandOnFinish) {
				commandOnFinish[i] = onFinish
				if !onFinish.Failed {
					states[i].State = ui.StateSuccess
					states[i].Msg = fmt.Sprintf("successfully retrieved status for %s", onFinish.Repo)
				} else {
					states[i].State = ui.StateError
					states[i].Msg = fmt.Sprintf("failed to pull status for %s", onFinish.Repo)
				}
			})
			return nil, nil
		}
	})

	feed := pool.Go()
	for res := range feed {
		if res.Error == nil {
			continue
		}
		util.Warning(res.Error.Error())
	}

	finishPrint()
	for _, onFinish := range commandOnFinish {
		if !onFinish.Failed {
			stdout := indent.String("        ", onFinish.RunConfig.StdOut.String())
			stderr := indent.String("        ", onFinish.RunConfig.StdErr.String())
			fmt.Printf("🟢 Status %s \n    Stdout:\n%s\n    StdErr:\n%s\n", onFinish.Repo, stdout, stderr)
		} else {
			stdout := indent.String("        ", onFinish.RunConfig.StdOut.String())
			stderr := indent.String("        ", onFinish.RunConfig.StdErr.String())
			fmt.Printf("🔴 Failed to get status %s \n    Stdout:\n%s\n    StdErr:\n%s\n", onFinish.Repo, stdout, stderr)
		}
	}
	return nil
}

func (cmd *StatusCommand) GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

func (cmd *StatusCommand) LoadConfiguration() (*types.GeeContext, error) {
	cwd, err := cmd.GetWorkingDirectory()
	if err != nil {
		return nil, err
	}
	return util.NewConfigHelper().LoadConfig(cwd)
}
