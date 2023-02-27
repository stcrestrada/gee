package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/pelletier/go-toml"
)

func GetHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, err
	}
	return err == nil, err
}

func ChDirGee() error {
	dir := fmt.Sprintf("%s/%s/", GetHomeDir(), ".gee")
	err := os.Chdir(dir)
	return err
}

func WriteRepoToConfig(ctx *GeeContext, cwd string) error {
	directories := strings.Split(cwd, "/")
	name := directories[len(directories)-1]
	config := ctx.Config

	if config.Repos == nil {
		config.Repos = []Repo{}
		repo := Repo{
			Name: name,
			Path: cwd,
		}
		config.Repos = append(config.Repos, repo)

	} else {
		if err := repoExists(config.Repos, name); err != nil {
			return err
		}
		config.Repos = append(config.Repos, Repo{
			Name: name,
			Path: cwd,
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

func WriteRepoLastCommitToJSON(repo string, lastCommit string) error {
	repos, err := readGeeJson()
	found := false
	if err != nil {
		return err
	}

	for _, r := range repos.GeeRepos {
		if r.Repo == repo {
			r.LastCommit = lastCommit
			found = true
			break
		}
		continue
	}

	if !found {
		repos.GeeRepos = append(repos.GeeRepos, GeeJSON{
			Repo:       repo,
			LastCommit: lastCommit,
		})
	}
	err = writeGeeJson(*repos)
	return err
}

func repoExists(repos []Repo, name string) error {
	var err error
	for _, repo := range repos {
		if repo.Name == name {
			errMsg := fmt.Sprintf("Repo %s already exists.", name)
			err = errors.New(errMsg)
			break
		}
		continue
	}
	return err
}

func getGeePath() string {
	return fmt.Sprintf("%s/%s", GetHomeDir(), "gee.toml")
}

func writeGeeJson(config GeeJsonConfig) error {
	byteValue, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return err
	}

	err = os.WriteFile("gee.json", byteValue, 0644)
	return err
}

func readGeeJson() (*GeeJsonConfig, error) {
	var config GeeJsonConfig
	err := ChDirGee()
	if err != nil {
		return nil, err
	}

	jsonFile, err := os.Open("gee.json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
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
				Branch: "main",
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
