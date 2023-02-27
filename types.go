package main

type Repo struct {
	// name of repo
	Name string `toml:"name"`
	// path of repo
	Path string `toml:"path"`

	// remote origin of repo
	Remote string `toml:"remote"`

	// branch of repo
	Branch string `toml:"branch"`
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
