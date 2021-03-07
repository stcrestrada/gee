package main

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
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
		initCmd := initCommand()
		app.Commands = append(app.Commands, initCmd)
		err = app.Run(os.Args)
		if err != nil {
			CheckIfError(err)
			return
		}
		return
	}
	config, err := setConfig(*configTree)
	if err != nil {
		addCmd := addCommand()
		Warning(fmt.Sprintf("%s \n", err))
		app.Commands = append(app.Commands, addCmd)
		err = app.Run(os.Args)
		if err != nil {
			CheckIfError(err)
			return
		}
		return
	}
	app.Commands = []*cli.Command{
		addCommand(),
		initCommand(),
		pullCommand(config),
		statusCommand(config),
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
