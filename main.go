package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gee/cmd"
	"gee/pkg/tui"
	"gee/pkg/util"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli/v2"
)

var (
	app     *cli.App
	version = "dev"
)

func main() {
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "Enable verbose logging",
		},
	}

	app.Before = func(c *cli.Context) error {
		verbose := c.Bool("verbose")
		util.SetVerbose(verbose)
		if verbose {
			util.VerboseLog("Verbose logging enabled")
		}
		return nil
	}

	app.Commands = []*cli.Command{
		cmd.AddCmd(),
		cmd.PullCmd(),
		cmd.StatusCmd(),
		cmd.RemoveCmd(),
		cmd.ExecCmd(),
	}

	// No subcommand → launch interactive TUI
	app.Action = func(c *cli.Context) error {
		cache := util.NewRepoCache()
		if _, err := cache.Load(); err != nil {
			return err
		}

		// One-time migration: import repos from gee.toml if cache is empty.
		if len(cache.All()) == 0 {
			migrateFromGeeToml(cache)
		}

		model := tui.NewAppModel(cache)
		p := tea.NewProgram(model, tea.WithAltScreen())
		_, err := p.Run()
		return err
	}

	// Run the CLI app
	err := app.Run(os.Args)
	if err != nil {
		switch err.(type) {
		case *util.InfoError:
			util.Info("%s", err.Error())
		case *util.WarningError:
			util.Warning("%s", err.Error())
		default:
			util.CheckIfError(err)
		}
	}
}

func init() {
	// Initialise a CLI app
	app = cli.NewApp()
	app.Name = "gee"
	app.Usage = "Manage git repos from one place"
	app.Version = version
}

// migrateFromGeeToml attempts a one-time import of repos from a gee.toml file.
// It searches from cwd upward. On success, each repo is added as pinned.
func migrateFromGeeToml(cache *util.RepoCache) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	configHelper := util.NewConfigHelper()
	geeTomlPath, err := configHelper.FindConfig(cwd, "gee.toml")
	if err != nil {
		return
	}

	tree, err := toml.LoadFile(geeTomlPath)
	if err != nil {
		return
	}

	type tomlConfig struct {
		Repos []struct {
			Name   string `toml:"name"`
			Path   string `toml:"path"`
			Remote string `toml:"remote"`
		} `toml:"repos"`
	}
	var cfg tomlConfig
	if err := tree.Unmarshal(&cfg); err != nil {
		return
	}

	imported := 0
	for _, r := range cfg.Repos {
		fullPath := filepath.Join(r.Path, r.Name)
		// Verify the repo still exists on disk.
		if info, err := os.Stat(filepath.Join(fullPath, ".git")); err != nil || !info.IsDir() {
			continue
		}
		cache.Add(util.CachedRepo{
			Name:         r.Name,
			Path:         fullPath,
			Remote:       r.Remote,
			Pinned:       true,
			DiscoveredAt: time.Now(),
		})
		imported++
	}

	if imported > 0 {
		cache.Save()
		fmt.Printf("Migrated %d repos from gee.toml → cache.json\n", imported)
	}
}
