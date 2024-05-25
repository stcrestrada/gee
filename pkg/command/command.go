package command

import (
	"gee/pkg/types"
	"github.com/urfave/cli/v2"
)

type Command interface {
	Run(c *cli.Context) error
	GetWorkingDirectory(string, error)
	LoadConfiguration(types.GeeContext, error)
}
