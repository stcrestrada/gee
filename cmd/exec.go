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
	"os/exec"
	"strings"
	"time"

	"github.com/stcrestrada/gogo/v3"
	"github.com/urfave/cli/v2"
)

type ExecCommand struct {
	RepoUtils *util.RepoUtils
}

func NewExecCommand() *ExecCommand {
	repoOp := command.GitRepoOperation{}
	return &ExecCommand{
		RepoUtils: util.NewRepoUtils(repoOp),
	}
}

func ExecCmd() *cli.Command {
	execCmd := NewExecCommand()
	return &cli.Command{
		Name:      "exec",
		Usage:     "Run a command in all repos",
		ArgsUsage: "<command>",
		Action: func(c *cli.Context) error {
			return execCmd.Run(c)
		},
	}
}

func (cmd *ExecCommand) Run(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return util.NewWarning("no command provided. usage: gee exec <command>")
	}

	startTime := time.Now()
	userCmd := strings.Join(c.Args().Slice(), " ")

	geeCtx, err := cmd.LoadConfiguration()
	if err != nil {
		return util.NewWarning(err.Error())
	}

	repos := geeCtx.Repos
	states := make([]*ui.SpinnerState, len(repos))
	results := make([]*execResult, len(repos))

	for i, repo := range repos {
		states[i] = &ui.SpinnerState{
			State: ui.StateLoading,
			Msg:   fmt.Sprintf("Running in %s", repo.Name),
		}
	}

	finishPrint := ui.PrintSpinnerStates(os.Stdout, states)

	concurrency := len(repos)
	pool := gogo.NewPool[struct{}](c.Context, concurrency, len(repos), func(ctx context.Context, i int) (struct{}, error) {
		repo := repos[i]
		fullPath := cmd.RepoUtils.FullPathWithRepo(repo.Path, repo.Name)

		var stdout, stderr bytes.Buffer
		sh := exec.Command("sh", "-c", userCmd)
		sh.Dir = fullPath
		sh.Stdout = &stdout
		sh.Stderr = &stderr

		err := sh.Run()
		failed := err != nil

		results[i] = &execResult{
			Repo:   repo.Name,
			Stdout: stdout.String(),
			Stderr: stderr.String(),
			Failed: failed,
		}

		if failed {
			states[i].State = ui.StateError
			states[i].Msg = fmt.Sprintf("failed in %s", repo.Name)
		} else {
			states[i].State = ui.StateSuccess
			states[i].Msg = fmt.Sprintf("finished in %s", repo.Name)
		}

		return struct{}{}, nil
	})

	for res := range pool.Go() {
		if res.Error == nil {
			continue
		}
		util.Warning(res.Error.Error())
	}

	finishPrint()
	fmt.Println()

	repoResults := make([]ui.RepoResult, len(results))
	for i, r := range results {
		repoResults[i] = ui.RepoResult{
			Name:   r.Repo,
			Stdout: r.Stdout,
			Stderr: r.Stderr,
			Failed: r.Failed,
		}
	}
	ui.RenderResults(fmt.Sprintf("$ %s", userCmd), repoResults, startTime)
	return nil
}

type execResult struct {
	Repo   string
	Stdout string
	Stderr string
	Failed bool
}

func (cmd *ExecCommand) GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

func (cmd *ExecCommand) LoadConfiguration() (*types.GeeContext, error) {
	cwd, err := cmd.GetWorkingDirectory()
	if err != nil {
		return nil, err
	}
	return util.NewConfigHelper().LoadConfig(cwd)
}
