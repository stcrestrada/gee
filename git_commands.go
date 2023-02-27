package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func (c *GitCommand) PullAll(branch string) ([]byte, error) {
	// in the event that the branch is prefixed with origin, we will parse to remove 'origin'
	var cmd *exec.Cmd
	if strings.HasPrefix(branch, "origin") {
		branchName := RemoveOriginFromBranchName(branch)
		cmd = exec.Command("git", "pull", "origin", fmt.Sprintf("%s", branchName))
	} else {
		cmd = exec.Command("git", "pull", "origin", fmt.Sprintf("%s", branch))
	}

	cmd.Dir = c.Dir
	output, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return output, errors.New(c.Stderr)
	}

	return output, nil
}

func (c *GitCommand) CurrentBranch() ([]byte, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = c.Dir
	output, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return output, errors.New(c.Stderr)
	}

	return output, nil
}

func (c *GitCommand) MainBranch() ([]byte, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "origin/HEAD")
	cmd.Dir = c.Dir
	output, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return output, errors.New(c.Stderr)
	}

	return output, err
}

func (c *GitCommand) Status() ([]byte, error) {
	cmd := exec.Command("git", "-c", "color.status=always", "status")
	cmd.Dir = c.Dir
	output, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return output, errors.New(c.Stderr)
	}

	return output, err
}

func (c *GitCommand) IsRepoClean() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = c.Dir

	output, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return false, errors.New(c.Stderr)
	}

	isClean := bytes.Equal(bytes.TrimSpace(output), []byte(""))
	return isClean, err
}

func (c *GitCommand) isClean() (bool, error) {
	cmd := exec.Command("git", "diff-index", "--quiet", "HEAD")
	cmd.Dir = c.Dir

	_, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return false, errors.New(c.Stderr)
	}

	return err == nil, err
}

func (c *GitCommand) isMainBranch() (bool, error) {
	mainBranch, err := c.MainBranch()
	if err != nil {
		return false, err
	}

	currBranch, err := c.CurrentBranch()
	if err != nil {
		return false, err
	}

	mBranch := []byte(RemoveOriginFromBranchName(string(mainBranch))) // returns string by convert back to byte array for check

	isSame := bytes.Equal(bytes.TrimSpace(mBranch), bytes.TrimSpace(currBranch))
	return isSame, err

}
func (c *GitCommand) Add() (bool, error) {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = c.Dir

	_, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return false, errors.New(c.Stderr)
	}

	return err == nil, err
}

func (c *GitCommand) Stash() (bool, error) {
	cmd := exec.Command("git", "stash")
	cmd.Dir = c.Dir

	_, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return false, errors.New(c.Stderr)
	}

	return err == nil, err
}

func (c *GitCommand) CheckoutToTmpBranch() (bool, error) {
	cmd := exec.Command("git", "checkout", "-b", "tmpmain")
	cmd.Dir = c.Dir

	_, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return false, errors.New(c.Stderr)
	}

	return err == nil, err
}

func (c *GitCommand) DeleteTmpBranch() (bool, error) {
	cmd := exec.Command("git", "branch", "-D", "tmpmain")
	cmd.Dir = c.Dir

	_, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return false, errors.New(c.Stderr)
	}

	return err == nil, err
}

func (c *GitCommand) StashApply() (bool, error) {
	cmd := exec.Command("git", "stash", "apply")
	cmd.Dir = c.Dir

	_, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return false, errors.New(c.Stderr)
	}

	return err == nil, err
}

func (c *GitCommand) AddAndStash() (bool, error) {
	_, err := c.Add()
	if err != nil {
		return false, err
	}

	_, err = c.Stash()
	if err != nil {
		return false, err
	}

	return true, err
}

func (c *GitCommand) LastCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = c.Dir // montezuma beach

	output, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return "", errors.New(c.Stderr)
	}

	return strings.TrimSpace(string(output)), err
}

func (c *GitCommand) AbortMergeConflict() (bool, error) {
	cmd := exec.Command("git", "reset", "--merge")
	cmd.Dir = c.Dir

	_, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return false, errors.New(c.Stderr)
	}

	return err == nil, err
}

func (c *GitCommand) Checkout(branch string) (bool, error) {
	var cmd *exec.Cmd
	if strings.HasPrefix(branch, "origin") {
		parseBranch := strings.Split(branch, "/")
		cmd = exec.Command("git", "checkout", fmt.Sprintf("%s", parseBranch[len(parseBranch)-1]))
	} else {
		cmd = exec.Command("git", "checkout", fmt.Sprintf("%s", branch))
	}
	cmd.Dir = c.Dir

	_, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return false, errors.New(c.Stderr)
	}

	return err == nil, err
}

func (c *GitCommand) RemoteOriginUrl() ([]byte, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = c.Dir

	output, err := cmd.CombinedOutput()

	c.Stderr = cmd.Stderr.(*bytes.Buffer).String()

	if len(c.Stderr) > 0 && err != nil {
		return output, errors.New(c.Stderr)
	}

	return output, nil
}

func (c *GitCommand) Clone(remoteUrl string, directory string) (io.ReadCloser, error) {
	cmd := exec.Command("git", "-C", directory, "clone", remoteUrl)
	cmd.Dir = c.Dir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}

	err = cmd.Start()
	fmt.Println("The command is running")
	if err != nil {
		fmt.Println(err)
	}

	// print the output of the subprocess
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		b := scanner.Bytes()
		m := scanner.Text()
		fmt.Println(m)
		fmt.Println(b)
	}
	cmd.Wait()
	fmt.Println("The command has finished")
	//c.Stderr = cmd.Stderr.(*bytes.Buffer).String()
	//
	//if len(c.Stderr) > 0 {
	//	return nil, errors.New(c.Stderr)
	//}
	return nil, nil
}
