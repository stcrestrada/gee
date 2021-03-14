package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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

	mainBranch, err := cmd.MainBranch()
	if err != nil {
		return nil, err
	}

	isClean, err := cmd.IsRepoClean()
	if err != nil {
		return nil, err
	}

	if !isClean {
		cmdResult, errr := ReapplyStashAndPull(repo, strings.TrimSpace(string(mainBranch)))
		if errr != nil {
			output := fmt.Sprintf("%s has unstashed changes unable to pull: %s \n", repo.Name, errr)
			return &CommandOutput{
				Repo:    repo.Name,
				Dir:     repo.Path,
				Output:  []byte(output),
				Warning: true,
			}, nil
		}
		return cmdResult, nil
	}

	pullOutput, err := cmd.PullAll(strings.TrimSpace(string(mainBranch)))
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
	var data GeeJsonConfig
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

		data.GeeRepos = append(data.GeeRepos, GeeJSON{})
		file, _ := json.MarshalIndent(data, "", " ")
		err = ioutil.WriteFile("gee.json", file, 0755)

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

func DeleteTmpBranch(repo Repo) (*CommandOutput, error) {
	cmd := GitCommand{
		Repo: repo.Name,
		Dir:  repo.Path,
	}
	_, err := cmd.DeleteTmpBranch()
	return &CommandOutput{
		Repo: repo.Name,
		Dir:  repo.Path,
	}, err
}

func ReapplyStashAndPull(repo Repo, branch string) (*CommandOutput, error) {
	cmd := GitCommand{
		Repo: repo.Name,
		Dir:  repo.Path,
	}

	// grab latest commit
	lastCommit, err := cmd.LastCommitHash()
	if err != nil {
		return nil, err
	}

	// step one
	// - stash changes
	_, err = cmd.AddAndStash()
	if err != nil {
		return nil, err
	}
	// step two
	// - check out to mastertemp branch
	_, err = cmd.CheckoutToTmpBranch()
	if err != nil {
		return nil, err
	}
	// step three
	// - pull down changes
	_, err = cmd.PullAll(branch)
	if err != nil {
		return nil, err
	}
	// step four
	// - apply stash changes
	_, err = cmd.StashApply()
	// - if there is a merge conflict
	if err != nil {
		// 	- abort merge
		_, _ = cmd.AbortMergeConflict()
		//	- checkout back out to master/main branch
		_, _ = cmd.Checkout(branch)
		// 	- delete tmp branch
		_, _ = cmd.DeleteTmpBranch()
		_, _ = cmd.StashApply()
		return nil, err
	}

	_, err = cmd.AddAndStash()
	if err != nil {
		return nil, err
	}

	_, _ = cmd.Checkout(branch)
	_, _ = cmd.DeleteTmpBranch()
	latestCommit, err := cmd.LastCommitHash()
	if latestCommit != latestCommit {
		err = WriteRepoLastCommitToJSON(repo.Name, lastCommit)
		if err != nil {
			return nil, err
		}
	}

	pullOutput, err := cmd.PullAll(branch)
	if err != nil {
		CheckIfError(err)
		return nil, err
	}
	_, _ = cmd.StashApply()

	return &CommandOutput{
		Repo:   repo.Name,
		Dir:    repo.Path,
		Output: pullOutput,
	}, err
}

func Rollback() {

}
