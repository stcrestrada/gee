package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"gee/pkg/command"
	"gee/pkg/types"
	"gee/pkg/ui"
	"gee/pkg/util"

	"github.com/stcrestrada/gogo/v3"
	"github.com/urfave/cli/v2"
)

func StatusCmd() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Git status of pinned repos (or current repo)",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "all",
				Usage: "Target all cached repos, not just pinned",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Show full git status output instead of summary",
			},
		},
		Action: func(c *cli.Context) error {
			startTime := time.Now()
			verbose := c.Bool("verbose")

			cache := util.NewRepoCache()
			if _, err := cache.Load(); err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cached := cache.LoadReposForCLI(cwd, c.Bool("all"))
			if len(cached) == 0 {
				fmt.Println("No repos found. Run gee add in a git repo to pin it.")
				return nil
			}

			repos := util.ToRepoSlice(cached)
			git := command.GitRepoOperation{}
			repoUtils := util.NewRepoUtils(git)

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
				fullPath := repoUtils.FullPathWithRepo(repo.Path, repo.Name)

				rc := &types.RunConfig{
					StdErr: &bytes.Buffer{},
					StdOut: &bytes.Buffer{},
				}

				if verbose {
					git.Status(repo.Name, fullPath, rc, func(onFinish *types.CommandOnFinish) {
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
					git.StatusPorcelain(repo.Name, fullPath, rc, func(onFinish *types.CommandOnFinish) {
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
							fullPath := repoUtils.FullPathWithRepo(repos[i].Path, repos[i].Name)
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
		},
	}
}
