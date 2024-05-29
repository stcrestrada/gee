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
	"strings"
)

type CloneCommand struct {
	Git       command.GitRepoOperation
	RepoUtils *util.RepoUtils
}

func NewCloneCommand() *CloneCommand {
	repoOp := command.GitRepoOperation{}
	return &CloneCommand{
		Git:       repoOp,
		RepoUtils: util.NewRepoUtils(repoOp),
	}
}

func CloneCmd() *cli.Command {
	cloneCmd := NewCloneCommand()
	return &cli.Command{
		Name:  "clone",
		Usage: "Git clone of all repos in gee.toml",
		Action: func(c *cli.Context) error {
			return cloneCmd.Run(c)
		},
	}
}

func (cmd *CloneCommand) Run(c *cli.Context) error {
	ctx, err := cmd.LoadConfiguration()
	if err != nil {
		return util.NewWarning(err.Error())
	}

	repos := ctx.Config.Repos

	states := make([]*ui.SpinnerState, len(repos))
	commandOnFinish := make([]*types.CommandOnFinish, len(repos))

	for i, repo := range repos {
		states[i] = &ui.SpinnerState{
			State: ui.StateLoading,
			Msg:   fmt.Sprintf("cloning %s", repo.Name),
		}
	}

	finishPrint := ui.PrintSpinnerStates(os.Stdout, states)

	concurrency := len(repos)
	pool := gogo.NewPool(concurrency, len(repos), func(i int) func() (interface{}, error) {
		repo := repos[i]
		state := states[i]
		return func() (interface{}, error) {
			err = cmd.RepoUtils.GetOrCreateDir(repo.Path)
			if err != nil {
				return nil, err
			}

			rc := &types.RunConfig{
				StdErr: &bytes.Buffer{},
				StdOut: &bytes.Buffer{},
			}

			cmd.Git.Clone(repo.Name, repo.Remote, repo.Path, rc, func(onFinish *types.CommandOnFinish) {
				if onFinish.Failed {
					if strings.Contains(rc.StdErr.String(), "already exists") {
						onFinish.Failed = false
						state.State = ui.StateSuccess
						state.Msg = fmt.Sprintf("%s is already cloned", repo.Name)
					} else {
						state.State = ui.StateError
						state.Msg = fmt.Sprintf("failed to clone %s", repo.Name)
					}

				} else {
					state.State = ui.StateSuccess
					state.Msg = fmt.Sprintf("finished cloning %s", repo.Name)
				}
				commandOnFinish[i] = onFinish
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
	os.Stdout.Write([]byte("\n\n"))
	for _, onFinish := range commandOnFinish {
		if onFinish.Failed {
			stdout := indent.String("        ", onFinish.RunConfig.StdOut.String())
			stderr := indent.String("        ", onFinish.RunConfig.StdErr.String())
			fmt.Printf("🟡 Failed to clone %s \n    Stdout:\n%s\n    StdErr:\n%s\n", onFinish.Repo, stdout, stderr)
		}
	}
	return nil
}

func (cmd *CloneCommand) GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

func (cmd *CloneCommand) LoadConfiguration() (*types.GeeContext, error) {
	cwd, err := cmd.GetWorkingDirectory()
	if err != nil {
		return nil, err
	}
	return util.NewConfigHelper().LoadConfig(cwd)
}
