package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gee/pkg/util"

	"github.com/urfave/cli/v2"
)

func AddCmd() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Pin the current repo (or add by path)",
		Action: func(c *cli.Context) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			target := cwd
			if c.Args().Len() > 0 {
				target, err = filepath.Abs(c.Args().First())
				if err != nil {
					return err
				}
			}

			// Verify it's a git repo.
			gitDir := filepath.Join(target, ".git")
			if info, err := os.Stat(gitDir); err != nil || !info.IsDir() {
				return util.NewWarning(fmt.Sprintf("%s is not a git repository", target))
			}

			cache := util.NewRepoCache()
			if _, err := cache.Load(); err != nil {
				return err
			}

			// Detect remote.
			remote := ""
			cmd := exec.Command("git", "-C", target, "config", "--get", "remote.origin.url")
			if out, err := cmd.Output(); err == nil {
				remote = strings.TrimSpace(string(out))
			}

			name := filepath.Base(target)
			cache.Add(util.CachedRepo{
				Name:         name,
				Path:         target,
				Remote:       remote,
				Pinned:       true,
				DiscoveredAt: time.Now(),
			})
			cache.Pin(target)

			if err := cache.Save(); err != nil {
				return err
			}

			return util.NewInfo(fmt.Sprintf("pinned %s", name))
		},
	}
}
