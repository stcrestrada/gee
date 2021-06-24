package main

type CommandOutput struct {
	Repo    string
	Dir     string
	Output  []byte
	Warning bool
}

type GitCommand struct {
	Repo string
	Dir  string
	Stderr string
}

type Repo struct {
	// name of repo
	Name string `toml:"name" validate:"required,min=1"`
	// path of repo
	Path string `toml:"path" validate:"required,min=1"`
}

type Config struct {
	Repos []Repo `toml:"repos" validate:"required,dive,required"`
}

type GeeJSON struct {
	Repo string `json:"repo,omitempty"`
	LastCommit string `json:"last_commit,omitempty"`
}

type GeeJsonConfig struct {
	GeeRepos []GeeJSON `json:"repos"`
}