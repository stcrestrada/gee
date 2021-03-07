package main

import (
	"github.com/stcrestrada/gogo"
	"github.com/urfave/cli/v2"
)

func initCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "initialize gee directory and toml file",
		Action: func(context *cli.Context) error {
			err := GeeInit()
			return err
		},
	}
}

func addCommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "add repo to gee.toml",
		Action: func(context *cli.Context) error {
			err := GeeAdd()
			return err
		},
	}
}

func pullCommand(config *Config) *cli.Command {
	return &cli.Command{
		Name:  "pull",
		Usage: "Git pull and update all repos",
		Action: func(c *cli.Context) error {
			concurrency := len(config.Repos)
			repos := config.Repos
			pool := gogo.NewPool(concurrency, len(repos), func(i int) func() (interface{}, error) {
				repo := repos[i]
				return func() (interface{}, error) {
					output, err := GeePullAll(repo)
					return output, err
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
						Info("Finished pulling %s \n", cmdOutput.Repo)
					}
					continue
				}
				Warning(res.Error.Error())
			}
			return nil
		},
	}
}

func statusCommand(config *Config) *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Git status of all repos",
		Action: func(c *cli.Context) error {
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
