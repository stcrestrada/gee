package main

import (
	"fmt"
	"os"
	"path"
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
		output := fmt.Sprintf("Skipping, cannot pull updates for repository -> %s. Currently not in main branch", repo.Name)

		return &CommandOutput{
			Repo:    repo.Name,
			Dir:     repo.Path,
			Output:  []byte(output),
			Warning: true,
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
			output := fmt.Sprintf("Error pulling %s -> %s \n", repo.Name, errr)
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

func GeeCreate(cwd string) error {
	geeFile := path.Join(cwd, "gee.toml")
	_, err := os.Stat(geeFile)
	if err != nil {
		return err
	}

	_, err = os.Create(geeFile)
	if err != nil {
		return err
	}

	return err
}

func GeeAdd(ctx *GeeContext, cwd string) error {
	exists, err := FileExists(".git")
	if !exists {
		Warning("Not a git initialize repo, .git dir does not exist")
		return err
	}
	err = WriteRepoToConfig(ctx, cwd)
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

func CleanupStaleBranches(repo Repo) (*Repo, error) {
	cmd := GitCommand{
		Repo: repo.Name,
		Dir:  repo.Path,
	}
	_, err := cmd.DeleteTmpBranch()

	return &repo, err
}

func GeeClone(repo Repo) (*CommandOutput, error) {
	cmd := GitCommand{
		Repo: repo.Name,
		Dir:  repo.Path,
	}
	_, err := cmd.Clone(repo.Remote, repo.Path)

	return &CommandOutput{
		Repo: repo.Name,
		Dir:  repo.Path,
		//Read: &output,
	}, err
}
