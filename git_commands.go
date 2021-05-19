package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func (c *GitCommand) PullAll(branch string) ([]byte, error) {
	cmd := exec.Command("git", "-c", "color.status=always", "pull", "origin", fmt.Sprintf("%s", branch))
	cmd.Dir = c.Dir
	output, err := cmd.Output()

	return output, err
}

func (c *GitCommand) CurrentBranch() ([]byte, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = c.Dir
	output, err := cmd.Output()

	return output, err
}

func (c *GitCommand) MainBranch() ([]byte, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "origin/HEAD")
	cmd.Dir = c.Dir
	output, err := cmd.Output()

	return output, err
}

func (c *GitCommand) Status() ([]byte, error) {
	cmd := exec.Command("git", "-c", "color.status=always", "status")
	cmd.Dir = c.Dir
	output, err := cmd.Output()

	return output, err
}

func (c *GitCommand) IsRepoClean() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = c.Dir

	output, err := cmd.Output()

	isClean := bytes.Equal(bytes.TrimSpace(output), []byte(""))
	return isClean, err
}

func (c *GitCommand) isClean() (bool, error) {
	cmd := exec.Command("git", "diff-index", "--quiet", "HEAD")
	cmd.Dir = c.Dir

	_, err := cmd.Output()

	return err == nil, err
}

func (c *GitCommand) isMainBranch() (bool, error) {
	mainBranch, err := c.MainBranch()
	if err != nil {
		return false, nil
	}

	currBranch, err := c.CurrentBranch()
	if err != nil {
		return false, nil
	}

	parseMainBranch := strings.Split(string(mainBranch), "/")
	branch := []byte(parseMainBranch[len(parseMainBranch) -1])

	isSame := bytes.Equal(bytes.TrimSpace(branch), bytes.TrimSpace(currBranch))
	return isSame, nil

}
func (c *GitCommand) Add() (bool, error) {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = c.Dir

	_, err := cmd.Output()

	return err == nil, err
}

func (c *GitCommand) Stash() (bool, error) {
	cmd := exec.Command("git", "stash")
	cmd.Dir = c.Dir

	_, err := cmd.Output()

	return err == nil, err
}

func (c *GitCommand) CheckoutToTmpBranch() (bool, error) {
	cmd := exec.Command("git", "checkout", "-b", "tmpmain")
	cmd.Dir = c.Dir

	_, err := cmd.Output()

	return err == nil, err
}

func (c *GitCommand) DeleteTmpBranch() (bool, error) {
	cmd := exec.Command("git", "branch", "-D", "tmpmain")
	cmd.Dir = c.Dir

	_, err := cmd.Output()

	return err == nil, err
}

func (c *GitCommand) StashApply() (bool, error) {
	cmd := exec.Command("git", "stash", "apply")
	cmd.Dir = c.Dir

	_, err := cmd.Output()

	return err == nil, err
}

func (c *GitCommand) AddAndStash() (bool, error) {
	_, err := c.Add()
	if err != nil {
		return false, err
	}

	_, err = c.Stash()
	if err != nil {
		return false, nil
	}

	return true, nil
}


func (c *GitCommand) LastCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = c.Dir  // montezuma beach

	output, err := cmd.Output()

	return strings.TrimSpace(string(output)), err
}

func (c *GitCommand) AbortMergeConflict() (bool, error) {
	cmd := exec.Command("git", "reset", "--merge")
	cmd.Dir = c.Dir

	_, err := cmd.Output()

	return err == nil, err
}

func (c *GitCommand) Checkout(branch string) (bool, error) {
	cmd := exec.Command("git", "checkout", fmt.Sprintf("%s", branch))
	cmd.Dir = c.Dir

	_, err := cmd.Output()

	return err == nil, err
}