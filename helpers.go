package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
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
		return false, nil
	}
	return err == nil, err
}

func ChDirGee() error {
	dir := fmt.Sprintf("%s/%s/", GetHomeDir(), ".gee")
	err := os.Chdir(dir)
	return err
}

func WriteRepoToConfig(config *Config, cd string, err error) error {
	directories := strings.Split(cd, "/")
	name := directories[len(directories)-1]
	geePath := getGeePath()

	if err != nil && config != nil {
		err = nil // set error to nil since it validating an error we do not want, this is weird workaround
		config.Repos = append(config.Repos, Repo{
			Name: name,
			Path: cd,
		})
		result, err := toml.Marshal(config)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(geePath, result, 0755)
		if err != nil {
			return err
		}
		Info("Successfully added repo in %s", cd)
	} else {
		if config != nil && len(config.Repos) > 0 {
			err = nil // set error to nil since it validating an error we do not want, this is weird workaround
			err = repoExists(config.Repos, name)
			if err != nil {
				return err
			}

			config.Repos = append(config.Repos, Repo{
				Name: name,
				Path: cd,
			})
			result, err := toml.Marshal(config)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(geePath, result, 0755)
			if err != nil {
				return err
			}
			Info("Successfully added repo in %s", cd)
		}
	}

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
			err = errors.New("Repo Already Added.")
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

	err = ioutil.WriteFile("gee.json", byteValue, 0644)
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
