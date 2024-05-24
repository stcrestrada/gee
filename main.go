package main

import (
	"gee/cmd"
	"gee/pkg/util"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/urfave/cli/v2"
)

var (
	app     *cli.App
	version = "dev"
)
var validate *validator.Validate

func main() {
	app.Commands = []*cli.Command{
		cmd.InitCmd(),
		cmd.AddCmd(),
		cmd.PullCmd(),
		cmd.StatusCmd(),
		cmd.CloneCmd(),
	}

	// Run the CLI app
	err := app.Run(os.Args)
	if err != nil {
		util.CheckIfError(err)
		return
	}
}

func init() {
	validate = validator.New()
	// Initialise a CLI app
	app = cli.NewApp()
	app.Name = "gee"
	app.Usage = "Gee gives user's control of git commands across repos without moving between them."
	app.Version = version
}
