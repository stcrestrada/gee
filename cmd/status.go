package cmd

import (
	"bytes"
	"context"
	"fmt"
	"gee/pkg/command"
	"gee/pkg/types"
	"gee/pkg/ui"
	"gee/pkg/util"
	"os"
	"time"

	"github.com/stcrestrada/gogo/v3"
	"github.com/urfave/cli/v2"
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
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Show full git status output instead of summary",
			},
		},
		Action: func(c *cli.Context) error {
			return statusCmd.Run(c)
		},
	}
}

func (cmd *StatusCommand) Run(c *cli.Context) error {
	startTime := time.Now()
	verbose := c.Bool("verbose")

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
	pool := gogo.NewPool[struct{}](c.Context, concurrency, len(repos), func(ctx context.Context, i int) (struct{}, error) {
		repo := repos[i]
		fullPath := cmd.RepoUtils.FullPathWithRepo(repo.Path, repo.Name)

		rc := &types.RunConfig{
			StdErr: &bytes.Buffer{},
			StdOut: &bytes.Buffer{},
		}

		if verbose {
			cmd.Git.Status(repo.Name, fullPath, rc, func(onFinish *types.CommandOnFinish) {
				commandOnFinish[i] = onFinish
				if !onFinish.Failed {
					states[i].State = ui.StateSuccess
					states[i].Msg = fmt.Sprintf("successfully retrieved status for %s", onFinish.Repo)
				} else {
					states[i].State = ui.StateError
					states[i].Msg = fmt.Sprintf("failed to get status for %s", onFinish.Repo)
				}
			})
		} else {
			cmd.Git.StatusPorcelain(repo.Name, fullPath, rc, func(onFinish *types.CommandOnFinish) {
				commandOnFinish[i] = onFinish
				if !onFinish.Failed {
					states[i].State = ui.StateSuccess
					states[i].Msg = fmt.Sprintf("successfully retrieved status for %s", onFinish.Repo)
				} else {
					states[i].State = ui.StateError
					states[i].Msg = fmt.Sprintf("failed to get status for %s", onFinish.Repo)
				}
			})
		}
		return struct{}{}, nil
	})

	for res := range pool.Go() {
		if res.Error == nil {
			continue
		}
		util.Warning("%s", res.Error)
	}

	finishPrint()
	fmt.Println()

	if verbose {
		repoResults := make([]ui.RepoResult, len(repos))
		for i, onFinish := range commandOnFinish {
			repoResults[i] = ui.RepoResult{
				Name:   repos[i].Name,
				Stdout: onFinish.RunConfig.StdOut.String(),
				Stderr: onFinish.RunConfig.StdErr.String(),
				Failed: onFinish.Failed,
			}
		}
		ui.RenderResults("", repoResults, startTime)
	} else {
		statusResults := make([]ui.RepoStatusResult, len(repos))
		for i, onFinish := range commandOnFinish {
			if onFinish.Failed {
				statusResults[i] = ui.RepoStatusResult{Name: repos[i].Name, Failed: true}
			} else {
				summary := ui.ParsePorcelainV2(onFinish.RunConfig.StdOut.String())
				if summary.Branch == "(detached)" {
					fullPath := cmd.RepoUtils.FullPathWithRepo(repos[i].Path, repos[i].Name)
					summary.State, summary.Progress = ui.DetectGitState(fullPath)
				}
				statusResults[i] = ui.RepoStatusResult{
					Name:    repos[i].Name,
					Summary: summary,
				}
			}
		}
		ui.RenderStatusTable(statusResults, startTime)
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
