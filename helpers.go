package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
)

const TEMPDIR = "/tmp"

func WriteCommitToTmpDir() {
	file, err := ioutil.TempFile("tmp", "cloudsynth.6771d542e3ab9b716908130a49d350db5123d39c.*")
	if err != nil {
		log.Fatal(err)
	}
	file2, err := ioutil.TempFile("tmp", "gee.6771d542e3ab9b716908130a49d350db5123d39c.*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer os.Remove(file2.Name())
}

func GetLastCommitFromTempFile(repo Repo) (string, error) {
	var files []string
	var lastCommit string

	os.Chdir(TEMPDIR)

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	err = filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", err
	}

	for _, file := range files {
		fileParts := strings.Split(file, ".")
		if len(files) != 3 {
			continue
		}
		repoName := fileParts[0]
		commit := fileParts[1]

		if repo.Name != repoName {
			continue
		}
		lastCommit = commit
		break
	}
	return lastCommit, nil
}

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
