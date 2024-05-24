package command

import "gee/pkg/types"

type Command interface {
	Run() error
	GetWorkingDirectory(string, error)
	LoadConfiguration(types.GeeContext, error)
}
