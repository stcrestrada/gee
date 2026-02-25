package cmd

import (
	"bytes"
	"context"
	"fmt"
	"gee/pkg/command"
	"gee/pkg/types"
	"gee/pkg/ui"
	"gee/pkg/util"
	"github.com/pborman/indent"
	"github.com/stcrestrada/gogo/v3"
	"github.com/urfave/cli/v2"
	"os"
)

type PullCommand struct {
	Git       command.GitRepoOperation
	RepoUtils *util.RepoUtils
}

func NewPullCommand() *PullCommand {
	repoOp := command.GitRepoOperation{}
	return &PullCommand{
		Git:       repoOp,
		RepoUtils: util.NewRepoUtils(repoOp),
	}
}

func PullCmd() *cli.Command {
	pullCmd := NewPullCommand()
	return &cli.Command{
		Name:  "pull",
		Usage: "Git pull and update all repos",
		Action: func(c *cli.Context) error {
			return pullCmd.Run(c)
		},
	}
}

func (cmd *PullCommand) Run(c *cli.Context) error {
	ctx, err := cmd.LoadConfiguration()
	if err != nil {
		util.Warning("Warning: %s \n", err)
		return err
	}

	repos := ctx.Repos

	states := make([]*ui.SpinnerState, len(repos))
	commandOnFinish := make([]*types.CommandOnFinish, len(repos))

	for i, repo := range repos {
		states[i] = &ui.SpinnerState{
			State: ui.StateLoading,
			Msg:   fmt.Sprintf("Pulling %s", repo.Name),
		}
	}

	finishPrint := ui.PrintSpinnerStates(os.Stdout, states)

	concurrency := len(repos)
	pool := gogo.NewPool[struct{}](c.Context, concurrency, len(repos), func(ctx context.Context, i int) (struct{}, error) {
		repo := repos[i]
		state := states[i]
		fullPath := cmd.RepoUtils.FullPathWithRepo(repo.Path, repo.Name)
		err = cmd.RepoUtils.GetOrCreateDir(repo.Path)
		if err != nil {
			return struct{}{}, err
		}

		rc := &types.RunConfig{
			StdErr: &bytes.Buffer{},
			StdOut: &bytes.Buffer{},
		}
		cmd.Git.Pull(repo.Name, fullPath, rc, func(onFinish *types.CommandOnFinish) {
			cmd.RepoUtils.HandlePullFinish(&repo, onFinish, state)
			commandOnFinish[i] = onFinish
		})
		return struct{}{}, nil
	})

	for res := range pool.Go() {
		if res.Error == nil {
			continue
		}
		util.Warning(res.Error.Error())
	}
	finishPrint()
	os.Stdout.Write([]byte("\n\n"))
	for _, onFinish := range commandOnFinish {
		if onFinish.Failed {
			stdout := indent.String("        ", onFinish.RunConfig.StdOut.String())
			stderr := indent.String("        ", onFinish.RunConfig.StdErr.String())
			fmt.Printf("🟡 Failed to pull %s \n    Stdout:\n%s\n    StdErr:\n%s\n", onFinish.Repo, stdout, stderr)
		}
	}
	return nil
}

func (cmd *PullCommand) GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

func (cmd *PullCommand) LoadConfiguration() (*types.GeeContext, error) {
	cwd, err := cmd.GetWorkingDirectory()
	if err != nil {
		return nil, err
	}
	return util.NewConfigHelper().LoadConfig(cwd)
}
