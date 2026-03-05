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

func PullCmd() *cli.Command {
	return &cli.Command{
		Name:  "pull",
		Usage: "Git pull pinned repos (or current repo)",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "all",
				Usage: "Target all cached repos, not just pinned",
			},
		},
		Action: func(c *cli.Context) error {
			startTime := time.Now()

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
					Msg:   fmt.Sprintf("Pulling %s", repo.Name),
				}
			}

			finishPrint := ui.PrintSpinnerStates(os.Stdout, states)

			concurrency := len(repos)
			pool := gogo.NewPool[struct{}](c.Context, concurrency, len(repos), func(ctx context.Context, i int) (struct{}, error) {
				repo := repos[i]
				state := states[i]
				fullPath := repoUtils.FullPathWithRepo(repo.Path, repo.Name)

				rc := &types.RunConfig{
					StdErr: &bytes.Buffer{},
					StdOut: &bytes.Buffer{},
				}
				git.Pull(repo.Name, fullPath, rc, func(onFinish *types.CommandOnFinish) {
					repoUtils.HandlePullFinish(&repo, onFinish, state)
					commandOnFinish[i] = onFinish
				})
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

			repoResults := make([]ui.RepoResult, len(repos))
			for i, onFinish := range commandOnFinish {
				repoResults[i] = ui.RepoResult{
					Name:   repos[i].Name,
					Stdout: onFinish.RunConfig.StdOut.String(),
					Stderr: onFinish.RunConfig.StdErr.String(),
					Failed: onFinish.Failed,
				}
			}
			ui.RenderResults("pull", repoResults, startTime)
			return nil
		},
	}
}
