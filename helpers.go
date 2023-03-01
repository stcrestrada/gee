package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pelletier/go-toml"
)

func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, err
	}
	return err == nil, err
}

func WriteRepoToConfig(ctx *GeeContext, cwd string) error {
	directories := strings.Split(cwd, "/")
	name := directories[len(directories)-1]
	config := ctx.Config

	if config.Repos == nil {
		config.Repos = []Repo{}
		repo := Repo{
			Name: name,
			Path: path.Dir(cwd),
		}
		config.Repos = append(config.Repos, repo)

	} else {
		if err := repoExists(config.Repos, cwd, name); err != nil {
			return err
		}
		config.Repos = append(config.Repos, Repo{
			Name: name,
			Path: path.Dir(cwd),
		})
	}

	result, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(ctx.ConfigFile, result, 0755)
	if err != nil {
		return err
	}
	Info("Successfully added repo in %s", cwd)
	return err
}

func repoExists(repos []Repo, cwd string, name string) error {
	var err error
	for _, repo := range repos {
		if repo.Name == name && repo.Path == cwd {
			errMsg := fmt.Sprintf("Repo %s already exists.", name)
			err = errors.New(errMsg)
			break
		}
		continue
	}
	return err
}

func RemoveOriginFromBranchName(branch string) string {
	// branch --> origin/master
	parseBranchName := strings.Split(branch, "/")
	grabLastItem := parseBranchName[len(parseBranchName)-1]
	return grabLastItem
}

// Create New Gee Context with Filler Config
func NewDummyGeeContext(cwd string) *GeeContext {
	config := Config{
		Repos: []Repo{
			Repo{
				Name:   "gee",
				Path:   cwd,
				Remote: "git@github.com:stcrestrada/gee.git",
			},
		},
	}

	geeConfigInfo := GeeConfigInfo{
		ConfigFile:     path.Join(cwd, "gee.toml"),
		ConfigFilePath: cwd,
	}

	return &GeeContext{
		GeeConfigInfo: geeConfigInfo,
		Config:        config,
	}
}

func InsertConfigIntoGeeToml(ctx *GeeContext) error {
	result, err := toml.Marshal(ctx.Config)
	if err != nil {
		return err
	}

	err = os.WriteFile(ctx.ConfigFile, result, 0755)
	if err != nil {
		return err
	}

	return nil
}

func FullPathWithRepo(repoPath string, repoName string) string {
	// get last directory in path
	lastDirName := repoPath[strings.LastIndex(repoPath, "/")+1:]
	if lastDirName == repoName {
		return repoPath
	}
	return path.Join(repoPath, repoName)
}

func GetOrCreateDir(path string) error {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		// path is a directory
		return nil
	}
	return os.MkdirAll(path, 0755)
}
