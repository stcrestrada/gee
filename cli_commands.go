package main

import (
	"bytes"
	"fmt"
	"github.com/pborman/indent"
	"github.com/stcrestrada/gogo"
	"github.com/urfave/cli/v2"
	"os"
)

func createCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "create gee.toml",
		Action: func(context *cli.Context) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			err = GeeCreate(cwd)
			if err == nil {
				Info("Created gee.toml in %s \n", cwd)
			} else {
				return err
			}

			// insert dummy data into gee.toml
			geeCtx := NewDummyGeeContext(cwd)
			err = InsertConfigIntoGeeToml(geeCtx)
			if err != nil {
				return err
			}
			return err
		},
	}
}

func addCommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "add repo to gee.toml",
		Action: func(context *cli.Context) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := LoadConfig(cwd)
			if err != nil {
				Warning("Warning: %s \n", err)
				return nil
			}

			err = GeeAdd(ctx, cwd)
			return err
		},
	}
}

func pullCommand() *cli.Command {
	return &cli.Command{
		Name:  "pull",
		Usage: "Git pull and update all repos",
		Action: func(c *cli.Context) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			ctx, err := LoadConfig(cwd)
			if err != nil {
				Warning("Warning: %s \n", err)
				return nil
			}

			config := ctx.Config
			concurrency := len(config.Repos)
			repos := config.Repos
			states := make([]*SpinnerState, len(repos))
			commandOnFinish := make([]*CommandOnFinish, len(repos))

			for i, repo := range repos {
				states[i] = &SpinnerState{
					State: StateLoading,
					Msg:   fmt.Sprintf("Pulling %s", repo.Name),
				}
			}

			finishPrint := PrintSpinnerStates(os.Stdout, states)

			pool := gogo.NewPool(concurrency, len(repos), func(i int) func() (interface{}, error) {
				repo := repos[i]
				state := states[i]
				return func() (interface{}, error) {
					rc := &RunConfig{
						StdErr: &bytes.Buffer{},
						StdOut: &bytes.Buffer{},
					}
					Pull(repo.Name, rc, func(onFinish *CommandOnFinish) {
						if onFinish.Failed {
							state.State = StateError
							state.Msg = fmt.Sprintf("Failed to pull %s", repo.Name)

						} else {
							state.State = StateSuccess
							state.Msg = fmt.Sprintf("Finished pulling %s", repo.Name)
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
				Warning(res.Error.Error())
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
		},
	}
}

func statusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Git status of all repos",
		Action: func(c *cli.Context) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			ctx, err := LoadConfig(cwd)
			if err != nil {
				Warning("Warning: %s \n", err)
				return nil
			}

			config := ctx.Config
			concurrency := len(config.Repos)
			repos := config.Repos
			pool := gogo.NewPool(concurrency, len(repos), func(i int) func() (interface{}, error) {
				repo := repos[i]
				return func() (interface{}, error) {
					output, err := GeeStatusAll(repo)
					return output, err
				}
			})
			feed := pool.Go()
			for res := range feed {
				if res.Error == nil {
					cmdOutput := res.Result.(*CommandOutput)
					Info("Status of %s \n", cmdOutput.Repo)
					println(string(cmdOutput.Output))
					continue
				}
				Warning(res.Error.Error())
			}
			return nil
		},
	}
}

func cloneCommand() *cli.Command {
	return &cli.Command{
		Name:  "clone",
		Usage: "Git clone of all repos in gee.toml",
		Action: func(c *cli.Context) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			ctx, err := LoadConfig(cwd)
			if err != nil {
				Warning("Warning: %s \n", err)
				return nil
			}

			config := ctx.Config
			concurrency := len(config.Repos)
			repos := config.Repos
			states := make([]*SpinnerState, len(repos))
			commandOnFinish := make([]*CommandOnFinish, len(repos))

			for i, repo := range repos {
				states[i] = &SpinnerState{
					State: StateLoading,
					Msg:   fmt.Sprintf("Cloning %s", repo.Name),
				}
			}

			finishPrint := PrintSpinnerStates(os.Stdout, states)

			pool := gogo.NewPool(concurrency, len(repos), func(i int) func() (interface{}, error) {
				repo := repos[i]
				state := states[i]
				return func() (interface{}, error) {
					rc := &RunConfig{
						StdErr: &bytes.Buffer{},
						StdOut: &bytes.Buffer{},
					}
					Clone(repo.Remote, repo.Path, repo.Name, rc, func(onFinish *CommandOnFinish) {
						if onFinish.Failed {
							state.State = StateError
							state.Msg = fmt.Sprintf("Failed to clone %s", repo.Name)

						} else {
							state.State = StateSuccess
							state.Msg = fmt.Sprintf("Finished cloning %s", repo.Name)
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
				Warning(res.Error.Error())
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
		},
	}
}
