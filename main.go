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
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "Enable verbose logging",
		},
	}

	app.Before = func(c *cli.Context) error {
		verbose := c.Bool("verbose")
		util.SetVerbose(verbose)
		if verbose {
			util.VerboseLog("Verbose logging enabled")
		}
		return nil
	}

	app.Commands = []*cli.Command{
		cmd.InitCmd(),
		cmd.AddCmd(),
		cmd.PullCmd(),
		cmd.StatusCmd(),
		cmd.CloneCmd(),
		cmd.RemoveCmd(),
	}

	// Run the CLI app
	err := app.Run(os.Args)
	if err != nil {
		switch err.(type) {
		case *util.InfoError:
			util.Info("%s", err.Error())
		case *util.WarningError:
			util.Warning("%s", err.Error())
		default:
			util.CheckIfError(err)
		}
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
