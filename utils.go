package main

import (
	"errors"
	"fmt"

	"github.com/pelletier/go-toml"
)

func loadToml() (*toml.Tree, error) {
	homeDir := GetHomeDir()
	filename := "gee.toml"
	geeFilePath := fmt.Sprintf("%s/%s", homeDir, filename)
	config, err := toml.LoadFile(geeFilePath)
	if err != nil {
		return nil, err
	}
	return config, err
}

func setConfig(config toml.Tree) (*Config, error) {
	conf := Config{}
	err := config.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}
	if len(conf.Repos) == 0 {
		err = errors.New("No repos. Use gee add to add repos to gee.toml")
		return &Config{
			Repos: nil,
		}, err
	}
	// validate that config has necessary fields
	err = validate.Struct(&conf)
	if err != nil {
		return nil, err
	}

	return &conf, err
}
