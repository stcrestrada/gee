package cmd

import (
	"fmt"
	"os"

	"gee/pkg/util"

	"github.com/urfave/cli/v2"
)

func RemoveCmd() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Unpin a repo",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repo",
				Aliases: []string{"r"},
				Usage:   "specify the repository name to unpin",
			},
		},
		Action: func(c *cli.Context) error {
			cache := util.NewRepoCache()
			if _, err := cache.Load(); err != nil {
				return err
			}

			repoName := c.String("repo")

			// If no name given, try to detect from cwd.
			if repoName == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				if found, ok := cache.FindByPath(cwd); ok {
					repoName = found.Name
				}
			}

			if repoName == "" {
				return util.NewWarning("please specify --repo <name> or run from inside a pinned repo")
			}

			// Find matching entry by name and unpin it.
			all := cache.All()
			var targetPath string
			var wasPinned bool
			for _, r := range all {
				if r.Name == repoName {
					targetPath = r.Path
					wasPinned = r.Pinned
					break
				}
			}

			if targetPath == "" {
				return util.NewWarning(fmt.Sprintf("%s not found in cache", repoName))
			}

			if wasPinned {
				cache.Unpin(targetPath)
			} else {
				cache.Remove(targetPath)
			}

			if err := cache.Save(); err != nil {
				return err
			}

			if wasPinned {
				return util.NewInfo(fmt.Sprintf("unpinned %s", repoName))
			}
			return util.NewInfo(fmt.Sprintf("removed %s from cache", repoName))
		},
	}
}
