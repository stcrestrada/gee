package main

import "io"

type CommandOutput struct {
	Repo      string
	RemoteUrl string
	Dir       string
	Output    []byte
	Read      *io.ReadCloser
	Warning   bool
}

type GitCommand struct {
	Repo   string
	Dir    string
	Stderr string
}

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

type GeeJSON struct {
	Repo       string `json:"repo,omitempty"`
	LastCommit string `json:"last_commit,omitempty"`
}

type GeeJsonConfig struct {
	GeeRepos []GeeJSON `json:"repos"`
}

type GeeConfigInfo struct {
	ConfigFile     string
	ConfigFilePath string
}

type GeeContext struct {
	GeeConfigInfo
	Config
}
