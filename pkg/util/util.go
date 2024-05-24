package util

import (
	"errors"
	"gee/pkg/types"
	"github.com/pelletier/go-toml"
	"os"
	"path"
)

type ConfigHelper struct {
	// Add fields here for shared state or dependencies if needed
}

func NewConfigHelper() *ConfigHelper {
	return &ConfigHelper{}
}

func (h *ConfigHelper) SetConfig(config toml.Tree) (*types.Config, error) {
	conf := types.Config{}
	err := config.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}
	return &conf, err
}

func (h *ConfigHelper) FindConfig(cwd string, fileName string) (string, error) {
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
	return h.FindConfig(parentDir, fileName)
}

func (h *ConfigHelper) LoadConfig(cwd string) (*types.GeeContext, error) {
	fileName := "gee.toml"
	configFile, err := h.FindConfig(cwd, fileName)
	if err != nil {
		return nil, err
	}
	geeConfig := &types.GeeConfigInfo{
		ConfigFile:     configFile,
		ConfigFilePath: path.Dir(configFile),
	}

	tomlConfig, err := toml.LoadFile(geeConfig.ConfigFile)
	if err != nil {
		return nil, err
	}

	conf, err := h.SetConfig(*tomlConfig)
	if err != nil {
		return nil, err
	}

	return &types.GeeContext{
		GeeConfigInfo: *geeConfig,
		Config:        *conf,
	}, nil
}
