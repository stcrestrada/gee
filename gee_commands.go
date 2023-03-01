package main

import (
	"errors"
	"os"
	"path"
)

func GeeCreate(cwd string) error {
	geeFile := path.Join(cwd, "gee.toml")
	_, err := os.Stat(geeFile)
	if err == nil {
		return errors.New("gee.toml already exists in this directory (or a parent directory)")
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
