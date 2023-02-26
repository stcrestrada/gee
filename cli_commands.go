package main

import (
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
			pool := gogo.NewPool(concurrency, len(repos), func(i int) func() (interface{}, error) {
				repo := repos[i]
				return func() (interface{}, error) {
					output, errr := GeePullAll(repo)
					return output, errr
				}
			})
			outputFeed := pool.Go()
			for res := range outputFeed {
				if res.Error == nil {
					cmdOutput := res.Result.(*CommandOutput)
					if cmdOutput.Warning {
						Warning(string(cmdOutput.Output))
					} else {
						Info("Pulling Repo %s \n", cmdOutput.Repo)
						println(string(cmdOutput.Output))
						Info("Finished Pulling %s \n", cmdOutput.Repo)
					}
					continue
				}
				Warning(res.Error.Error())
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
