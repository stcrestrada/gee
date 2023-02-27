package main

import (
	"errors"
	"github.com/pelletier/go-toml"
	"os"
	"path"
)

func setConfig(config toml.Tree) (*Config, error) {
	conf := Config{}
	err := config.Unmarshal(&conf);
	if err != nil {
		return nil, err
	}

	return &conf, err
}

func FindConfig(cwd string, fileName string) (string, error) {
	// if cwd is empty, we are in the root directory
	if cwd == "" {
		return "", errors.New("gee.toml not found")
	}
	workingPath := path.Join(cwd, fileName)
	if _, err := os.Stat(workingPath); err == nil {
		return workingPath, nil
	}
	parentDir := path.Dir(cwd)
	if parentDir == cwd {
		return "", errors.New("gee.toml not found")
	}
	return FindConfig(parentDir, fileName)
}

func LoadConfig(cwd string) (*GeeContext, error) {
	var geeConfig *GeeConfigInfo

	fileName := "gee.toml"
	configFile, err := FindConfig(cwd, fileName)
	if err != nil {
		return nil, err
	}
	geeConfig = &GeeConfigInfo{
		ConfigFile:     configFile,
		ConfigFilePath: path.Dir(configFile),
	}

	tomlConfig, err := toml.LoadFile(geeConfig.ConfigFile)
	if err != nil {
		return nil, err
	}

	conf, err := setConfig(*tomlConfig)
	if err != nil {
		return nil, err
	}

	return &GeeContext{
		*geeConfig,
		*conf,
	}, nil

}
