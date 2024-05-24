package types

import "bytes"

type Repo struct {
	// name of repo
	Name string `toml:"name"`
	// path of repo
	Path string `toml:"path"`
	// remote origin of repo
	Remote string `toml:"remote"`
}

type Config struct {
	Repos []Repo `toml:"repos" validate:"required,dive,required"`
}

type GeeConfigInfo struct {
	ConfigFile     string
	ConfigFilePath string
}

type GeeContext struct {
	GeeConfigInfo
	Config
}

type RunConfig struct {
	StdOut *bytes.Buffer
	StdErr *bytes.Buffer
}

type CommandOnFinish struct {
	Repo      string
	Path      string
	RunConfig *RunConfig
	Failed    bool
	Error     error
}
