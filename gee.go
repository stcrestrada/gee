package main

import (
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/stcrestrada/gogo"
	"github.com/urfave/cli/v2"
)

var (
	app *cli.App
)
var validate *validator.Validate

func main() {
	configTree, err := loadToml()
	if err != nil {
		Warning("%s", err)
		Info("Run gee init")
		return
	}
	config, err := setConfig(*configTree)
	if err != nil {
		CheckIfError(err)
		return
	}
	app.Commands = []*cli.Command{
		{
			Name:  "init",
			Usage: "initialize gee directory and toml file",
			Action: func(context *cli.Context) error {
				err := GeeInit()
				return err
			},
		},
		{
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
		},
		{
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
		},
	}

	// Run the CLI app
	err = app.Run(os.Args)
	if err != nil {
		CheckIfError(err)
		return
	}
}

func init() {
	validate = validator.New()
	// Initialise a CLI app
	app = cli.NewApp()
	app.Name = "gee"
	app.Usage = "Gee gives control of git across repos without changing directories"
	app.Version = "0.0.0"
}
