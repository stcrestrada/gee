package util

import (
	"bytes"
	"errors"
	"fmt"
	"gee/pkg/command"
	"gee/pkg/types"
	"gee/pkg/ui"
	"github.com/pelletier/go-toml"
	"os"
	"path"
	"strings"
)

type RepoUtils struct {
	RepoOp command.GitRepoOperation
}

func NewRepoUtils(repoOp command.GitRepoOperation) *RepoUtils {
	return &RepoUtils{
		RepoOp: repoOp,
	}
}

// GeeCreate creates a gee.toml file in the specified directory
func (r *RepoUtils) GeeCreate(cwd string) error {
	geeFile := path.Join(cwd, "gee.toml")
	_, err := os.Stat(geeFile)
	if err == nil {
		return errors.New("gee.toml already exists in this directory (or a parent directory)")
	}

	_, err = os.Create(geeFile)
	if err != nil {
		return err
	}

	return nil
}

// GeeAdd adds repository information to the gee.toml configuration
func (r *RepoUtils) GeeAdd(ctx *types.GeeContext, cwd string) error {
	exists, err := r.FileExists(".git")
	if !exists {
		return NewWarning("Not a git initialized repo, .git dir does not exist")
	}
	paths := strings.Split(cwd, "/")
	name := paths[len(paths)-1]
	err = r.RepoExists(ctx.Repos, cwd, name)
	if err != nil {
		return NewWarning(fmt.Sprintf("%s already in gee.toml", name))
	}
	err = r.WriteRepoToConfig(ctx, cwd)
	return err
}

// HandlePullFinish handles the finish of a pull operation
func (r *RepoUtils) HandlePullFinish(repo *types.Repo, onFinish *types.CommandOnFinish, state *ui.SpinnerState) {
	switch {
	case onFinish.Failed && strings.Contains(onFinish.RunConfig.StdErr.String(), "No such file or directory") && repo.Remote != "":
		state.Msg = fmt.Sprintf("Cloning instead...")
		r.RepoOp.Clone(repo.Name, repo.Remote, repo.Path, onFinish.RunConfig, r.HandleCloneFinish(repo, state))
	case onFinish.Failed:
		state.State = ui.StateError
		state.Msg = fmt.Sprintf("Failed to pull %s", repo.Name)
	default:
		state.State = ui.StateSuccess
		state.Msg = fmt.Sprintf("%s already up to date", onFinish.Repo)
	}
}

// HandleCloneFinish returns a function to handle the finish of a clone operation
func (r *RepoUtils) HandleCloneFinish(repo *types.Repo, state *ui.SpinnerState) func(onFinish *types.CommandOnFinish) {
	return func(onFinish *types.CommandOnFinish) {
		switch {
		case onFinish.Failed && strings.Contains(onFinish.RunConfig.StdErr.String(), "already exists"):
			state.State = ui.StateSuccess
			state.Msg = fmt.Sprintf("Already cloned %s", repo.Name)
		case onFinish.Failed:
			state.State = ui.StateError
			state.Msg = fmt.Sprintf("Failed to clone %s", repo.Name)
		default:
			state.State = ui.StateSuccess
			state.Msg = fmt.Sprintf("Finished cloning %s", repo.Name)
		}
	}
}

// WriteRepoToConfig writes the repository information to the configuration
func (r *RepoUtils) WriteRepoToConfig(ctx *types.GeeContext, cwd string) error {
	directories := strings.Split(cwd, "/")
	name := directories[len(directories)-1]
	config := ctx.Config

	// Create a new RunConfig
	rc := &types.RunConfig{
		StdOut: new(bytes.Buffer),
		StdErr: new(bytes.Buffer),
	}

	// Get the remote URL using the GetRemoteURL function
	var remote string
	r.RepoOp.GetRemoteURL(name, cwd, rc, func(onFinish *types.CommandOnFinish) {
		if onFinish.Failed {
			// Handle the error
			fmt.Println("Error getting remote URL:", onFinish.Error)
			return
		}
		remote = strings.TrimSpace(onFinish.RunConfig.StdOut.String())
	})

	if config.Repos == nil {
		config.Repos = []types.Repo{}
	}

	if err := r.RepoExists(config.Repos, cwd, name); err != nil {
		return err
	}

	config.Repos = append(config.Repos, types.Repo{
		Name:   name,
		Path:   path.Dir(cwd),
		Remote: remote,
	})

	result, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(ctx.ConfigFile, result, 0755)
	if err != nil {
		return err
	}
	return NewInfo(fmt.Sprintf("successfully added %s", name))
}

// RepoExists checks if the repository already exists in the provided list
func (r *RepoUtils) RepoExists(repos []types.Repo, cwd string, name string) error {
	for _, repo := range repos {
		pathToRepo := path.Join(repo.Path, repo.Name)
		if pathToRepo == cwd && repo.Name == name {
			return NewWarning(fmt.Sprintf("repository %s already exists at %s", name, cwd))
		}
	}
	return nil
}

// FileExists checks if a file with the given name exists
func (r *RepoUtils) FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// NewDummyGeeContext creates a new dummy GeeContext
func (r *RepoUtils) NewDummyGeeContext(cwd string) *types.GeeContext {
	return &types.GeeContext{
		GeeConfigInfo: types.GeeConfigInfo{
			ConfigFile:     path.Join(cwd, "gee.toml"),
			ConfigFilePath: cwd,
		},
		Config: types.Config{
			Repos: []types.Repo{
				{Name: "gee", Path: cwd, Remote: "git@github.com:stcrestrada/gee.git"},
			},
		},
	}
}

// InsertConfigIntoGeeToml inserts the configuration into the gee.toml file
func (r *RepoUtils) InsertConfigIntoGeeToml(ctx *types.GeeContext) error {
	f, err := os.OpenFile(ctx.ConfigFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(ctx.Config)
}

// FullPathWithRepo constructs the full path with the repository name
func (r *RepoUtils) FullPathWithRepo(repoPath string, repoName string) string {
	if path.Base(repoPath) != repoName {
		return path.Join(repoPath, repoName)
	}
	return repoPath
}

// GetOrCreateDir checks if the directory exists or creates it if necessary
func (r *RepoUtils) GetOrCreateDir(dirPath string) error {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return NewWarning(fmt.Sprintf("path %s exists and is not a directory", dirPath))
	}
	return nil
}
