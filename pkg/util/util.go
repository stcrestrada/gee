package util

import (
	"errors"
	"gee/pkg/types"
	"github.com/pelletier/go-toml"
	"os"
	"path"
	"path/filepath"
)

type ConfigHelper struct {
	// Add fields here for shared state or dependencies if needed
}

func NewConfigHelper() *ConfigHelper {
	return &ConfigHelper{}
}

func (h *ConfigHelper) ParseConfig(config toml.Tree) (*types.Config, error) {
	conf := types.Config{}
	err := config.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}
	return &conf, err
}

// SaveConfig writes the configuration to the gee.toml file
func (h *ConfigHelper) SaveConfig(dir string, config *types.GeeContext) error {
	configFilePath := filepath.Join(dir, "gee.toml")
	configTree, err := toml.Marshal(*config)
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath, configTree, 0644)
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

	conf, err := h.ParseConfig(*tomlConfig)
	if err != nil {
		return nil, err
	}
	VerboseLog("loaded gee.toml from %s", geeConfig.ConfigFilePath)
	return &types.GeeContext{
		GeeConfigInfo: *geeConfig,
		Config:        *conf,
	}, nil
}
