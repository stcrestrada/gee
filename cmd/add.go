package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gee/pkg/util"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v2"
)

func AddCmd() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Pin the current repo (or add by path)",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "all",
				Usage: "Pin all git repos found in the current directory",
			},
			&cli.BoolFlag{
				Name:  "all-select",
				Usage: "Interactively select which repos to pin from the current directory",
			},
		},
		Action: func(c *cli.Context) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cache := util.NewRepoCache()
			if _, err := cache.Load(); err != nil {
				return err
			}

			if c.Bool("all") {
				return addAll(cwd, cache)
			}
			if c.Bool("all-select") {
				return addAllSelect(cwd, cache)
			}

			return addSingle(c, cwd, cache)
		},
	}
}

// addSingle pins a single repo (the cwd or a path argument).
func addSingle(c *cli.Context, cwd string, cache *util.RepoCache) error {
	target := cwd
	if c.Args().Len() > 0 {
		var err error
		target, err = filepath.Abs(c.Args().First())
		if err != nil {
			return err
		}
	}

	if !isGitRepo(target) {
		return util.NewWarning(fmt.Sprintf("%s is not a git repository", target))
	}

	pinRepo(target, cache)
	if err := cache.Save(); err != nil {
		return err
	}
	return util.NewInfo(fmt.Sprintf("pinned %s", filepath.Base(target)))
}

// addAll discovers all child git repos and pins them all.
func addAll(dir string, cache *util.RepoCache) error {
	repos := discoverChildRepos(dir)
	if len(repos) == 0 {
		fmt.Println("No git repos found in", dir)
		return nil
	}

	pinned := 0
	for _, repoPath := range repos {
		pinRepo(repoPath, cache)
		pinned++
	}

	if err := cache.Save(); err != nil {
		return err
	}
	fmt.Printf("Pinned %d repos\n", pinned)
	return nil
}

// addAllSelect discovers child git repos and shows an interactive picker.
func addAllSelect(dir string, cache *util.RepoCache) error {
	repos := discoverChildRepos(dir)
	if len(repos) == 0 {
		fmt.Println("No git repos found in", dir)
		return nil
	}

	// Build options for the multi-select.
	options := make([]huh.Option[string], len(repos))
	for i, repoPath := range repos {
		options[i] = huh.NewOption(filepath.Base(repoPath), repoPath)
	}

	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select repos to pin").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if len(selected) == 0 {
		fmt.Println("No repos selected")
		return nil
	}

	for _, repoPath := range selected {
		pinRepo(repoPath, cache)
	}

	if err := cache.Save(); err != nil {
		return err
	}
	fmt.Printf("Pinned %d repos\n", len(selected))
	return nil
}

// discoverChildRepos finds immediate child directories that contain .git.
func discoverChildRepos(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var repos []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		child := filepath.Join(dir, e.Name())
		if isGitRepo(child) {
			repos = append(repos, child)
		}
	}
	return repos
}

// isGitRepo checks if a directory contains a .git subdirectory.
func isGitRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && info.IsDir()
}

// pinRepo adds a repo to the cache as pinned, detecting its remote URL.
func pinRepo(repoPath string, cache *util.RepoCache) {
	remote := ""
	cmd := exec.Command("git", "-C", repoPath, "config", "--get", "remote.origin.url")
	if out, err := cmd.Output(); err == nil {
		remote = strings.TrimSpace(string(out))
	}

	name := filepath.Base(repoPath)
	cache.Add(util.CachedRepo{
		Name:         name,
		Path:         repoPath,
		Remote:       remote,
		Pinned:       true,
		DiscoveredAt: time.Now(),
	})
	cache.Pin(repoPath)
}
