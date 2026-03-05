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
		&cli.BoolFlag{
			Name:  "init",
			Usage: "Print shell integration function to stdout",
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

	// No subcommand → launch interactive TUI (or handle --init)
	app.Action = func(c *cli.Context) error {
		if c.Bool("init") {
			fmt.Print(shellInitScript())
			return nil
		}

		cache := util.NewRepoCache()
		if _, err := cache.Load(); err != nil {
			return err
		}

		// One-time migration: import repos from gee.toml if cache is empty.
		if len(cache.All()) == 0 {
			migrateFromGeeToml(cache)
		}

		model := tui.NewAppModel(cache)
		p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		// Teleport: if the user pressed Enter on a repo, print the path to stdout.
		// The shell wrapper function (from gee --init) will cd into it.
		if m, ok := finalModel.(tui.AppModel); ok && m.SelectedPath != "" {
			fmt.Fprintln(os.Stdout, m.SelectedPath)
		}

		return nil
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

// shellInitScript returns the shell function wrapper for the user's shell.
// Users add `eval "$(gee --init)"` to their .zshrc/.bashrc.
func shellInitScript() string {
	shell := filepath.Base(os.Getenv("SHELL"))
	switch shell {
	case "fish":
		return fishInit()
	default:
		return bashZshInit()
	}
}

func bashZshInit() string {
	return `gee() {
  if [ $# -eq 0 ]; then
    local result
    result="$(command gee)"
    if [ -n "$result" ] && [ -d "$result" ]; then
      cd "$result" || return
    fi
  else
    command gee "$@"
  fi
}
`
}

func fishInit() string {
	return `function gee
  if test (count $argv) -eq 0
    set -l result (command gee)
    if test -n "$result" -a -d "$result"
      cd $result
    end
  else
    command gee $argv
  end
end
`
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
		fmt.Fprintf(os.Stderr, "Migrated %d repos from gee.toml → cache.json\n", imported)
	}
}
