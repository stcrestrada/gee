package main

import (
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"github.com/thoas/go-funk"
	"os"
	"path"
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
	//if len(conf.Repos) == 0 {
	//	err = errors.New("No repos. Use gee add to insert repos to gee.toml")
	//	return &Config{
	//		Repos: nil,
	//	}, err
	//}
	//// validate that config has necessary fields
	//err = validate.Struct(&conf)
	//if err != nil {
	//	return nil, err
	//}

	return &conf, err
}

func containsArg(args []string, elem string) bool {
	return funk.Contains(args, elem)
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
