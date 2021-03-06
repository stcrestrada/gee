package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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
		log.Fatal( err )
	}
	return usr.HomeDir
}

func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err != nil, err
}