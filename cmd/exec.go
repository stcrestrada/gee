package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gee/pkg/command"
	"gee/pkg/ui"
	"gee/pkg/util"

	"github.com/stcrestrada/gogo/v3"
	"github.com/urfave/cli/v2"
)

func ExecCmd() *cli.Command {
	return &cli.Command{
		Name:      "exec",
		Usage:     "Run a command in pinned repos (or current repo)",
		ArgsUsage: "<command>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "all",
				Usage: "Target all cached repos, not just pinned",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Args().Len() == 0 {
				return util.NewWarning("no command provided. usage: gee exec <command>")
			}

			startTime := time.Now()
			userCmd := strings.Join(c.Args().Slice(), " ")

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
				fullPath := repoUtils.FullPathWithRepo(repo.Path, repo.Name)

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
				util.Warning("%s", res.Error)
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
		},
	}
}

type execResult struct {
	Repo   string
	Stdout string
	Stderr string
	Failed bool
}
