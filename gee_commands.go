package main

import (
	"fmt"
	"os"
)

func GeeStatusAll(repo Repo) (*CommandOutput, error) {
	cmd := GitCommand{
		Repo: repo.Name,
		Dir:  repo.Path,
	}
	statusOutput, err := cmd.Status()
	if err != nil {
		CheckIfError(err)
		return nil, err
	}

	return &CommandOutput{
		Repo:   repo.Name,
		Dir:    repo.Path,
		Output: statusOutput,
	}, nil
}

func GeePullAll(repo Repo) (*CommandOutput, error) {
	cmd := GitCommand{
		Repo: repo.Name,
		Dir:  repo.Path,
	}
	isSameBranch, err := cmd.isMainBranch()
	if err != nil {
		return nil, err
	}

	if !isSameBranch {
		output := fmt.Sprintf("skipping, cannot update repo, %s. Currently not in main branch", repo.Name)

		return &CommandOutput{
			Repo:   repo.Name,
			Dir:    repo.Path,
			Output: []byte(output),
		}, nil
	}

	isClean, err := cmd.IsRepoClean()
	if err != nil {
		return nil, err
	}

	if !isClean {
		output := fmt.Sprintf("%s has unstashed changes \n", repo.Name)
		return &CommandOutput{
			Repo:    repo.Name,
			Dir:     repo.Path,
			Output:  []byte(output),
			Warning: true,
		}, nil
	}

	pullOutput, err := cmd.PullAll()
	if err != nil {
		CheckIfError(err)
		return nil, err
	}
	return &CommandOutput{
		Repo:   repo.Name,
		Dir:    repo.Path,
		Output: pullOutput,
	}, err
}

func GeeInit() error {
	// create gee.toml
	// create .gee/gee.json
	homeDir := GetHomeDir()

	err := os.Chdir(homeDir)
	if err != nil {
		return err
	}
	exists, err := FileExists("gee.toml")
	if err != nil {
		return err
	}

	if !exists {
		_, err = os.Create("gee.toml")
		if err != nil {
			return err
		}
	}

	cd, err := os.Getwd()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s", cd, ".gee/")
	exists, err = FileExists(path)
	if !exists {
		err = os.Mkdir(".gee", 0755)
		if err != nil {
			return err
		}

		err = os.Chdir(".gee")

		_, err = os.Create("gee.json")

	}
	if err == nil {
		Info("Gee initialized. You can  now add repos to gee.toml, located at %s. \n", homeDir)
		Info("To automate adding repos to gee.toml, use gee add inside of a git initialized repo.")
		return err
	}
	return err
}

func GeeAdd() error {
	cd, err := os.Getwd()
	exists, err := FileExists(".git")
	if !exists {
		Warning("Not a git initialize repo, .git dir does not exist")
		return err
	}
	config, err := loadToml()
	if err != nil {
		return err
	}
	conf, err := setConfig(*config)
	err = WriteRepoToConfig(conf, cd, err)
	return err
}
