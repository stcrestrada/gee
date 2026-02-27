# Gee: Efficiently Manage Multiple Git Repositories

Gee is a powerful tool for managing multiple git repositories. It features a full-screen interactive TUI (Terminal User Interface) with live status updates, Vim-style navigation, and one-key actions — plus a traditional CLI for scripting and automation. All operations run concurrently using a worker pool for speed.

<!-- TODO: Screenshot — full dashboard view showing several repos with mixed status (clean, modified, ahead/behind) -->

## Features

- **Interactive Dashboard**: Run `gee` to launch a K9s-style full-screen TUI with live-updating repo status
- **Vim Navigation**: `j`/`k` to move, `g`/`G` to jump, `/` to filter repos by name
- **One-Key Actions**: `p` to pull, `e` to exec a command, `Enter` to open a shell in any repo
- **Remote Discovery**: Press `d` to browse your GitHub/GitLab repos, multi-select, and batch-clone them
- **CLI Mode**: All traditional commands (`gee status`, `gee pull`, `gee exec`, etc.) still work for scripting
- **Concurrent Operations**: Every operation runs in parallel across all repos

## Installation

Install Gee using Homebrew:
```
brew tap stcrestrada/gee https://github.com/stcrestrada/gee

brew install gee
```

### Upgrade

```
brew update
brew upgrade gee
```

### Uninstall
```
brew uninstall gee
```

## Quick Start

```shell
# Create a config file
gee init

# Add repos from inside each repo directory
cd path/to/repo && gee add

# Launch the interactive dashboard
gee
```

## Interactive TUI

Run `gee` with no arguments to launch the interactive dashboard.

<!-- TODO: Screenshot — dashboard showing the header, repo table with branch/sync/changes columns, and help bar at the bottom -->

### Dashboard

The dashboard shows all your repos in a live-updating table with:
- Branch name (with rebase/merge/cherry-pick state detection)
- Sync status (ahead/behind remote)
- Change counts (staged, modified, untracked, conflicts)

Status refreshes automatically every 5 seconds and after every action.

### Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` | Move cursor down / up |
| `g` / `G` | Jump to first / last repo |
| `p` | Pull the selected repo |
| `P` | Pull all visible repos |
| `e` | Open exec prompt — run any shell command in the selected repo |
| `Enter` | Open a sub-shell in the selected repo's directory |
| `r` | Manually refresh status |
| `/` | Filter repos by name |
| `d` | Open the Discovery view (requires `gh` or `glab`) |
| `q` | Quit |

<!-- TODO: Screenshot — dashboard with the exec prompt open (showing "exec> " at the bottom) -->

### Discovery

Press `d` to open the Discovery view, which lists your remote repositories from GitHub or GitLab.

<!-- TODO: Screenshot — discovery view showing a list of remote repos with some selected (checkmarks) -->

| Key | Action |
|-----|--------|
| `j` / `k` | Move cursor down / up |
| `Space` | Toggle selection on the current repo |
| `Enter` | Clone all selected repos and add them to `gee.toml` |
| `Esc` | Return to the dashboard |

Discovery requires `gh` (GitHub CLI) or `glab` (GitLab CLI) to be installed. If neither is available, the `d` key is hidden from the help bar. GitHub is preferred when both are present.

## CLI Commands

All commands also work as traditional CLI subcommands for scripting and CI pipelines.

### Initialize Configuration
Create an initial `gee.toml` configuration file:

```
gee init
```

### Add Repository
Add a repository to the `gee.toml` file. Must be run from inside a git repository:
```
cd path/to/repo
gee add
```

### Check Status
Show a compact summary of all repos:
```
gee status
```

Example output:
```
✓  api        main   ↑2  ~3 modified  ?1 untracked
✓  frontend   dev    +1 staged  ~2 modified
✓  gee        main   clean
```

For full `git status` output, use verbose mode:
```
gee status --verbose
```

### Pull Changes
Pull changes for all repositories:
```
gee pull
```

### Clone Repositories
Clone all repositories listed in your `gee.toml` file:
```shell
gee clone
```

### Execute Commands
Run any command across all repos concurrently:
```shell
gee exec git push origin main
gee exec git stash
gee exec npm install
gee exec "git stash && git pull && git stash pop"
```

### Remove Repository
Automatically detect from the current directory:
```
gee remove
```
Specify repository name using flag:
```shell
gee remove -r repo_name
```
```shell
gee remove --repo repo_name
```

## Configuration

Manually configure `gee.toml`:
```toml
ConfigFile = "/path/to/gee/gee.toml"
ConfigFilePath = "/path/to/gee"

[[repos]]
name = "repo_name"
path = "/path/to/repo"
remote = "https://github.com/user/repo.git"
```

### Finding gee.toml
Gee searches for `gee.toml` in the current and parent directories. If no configuration file is found, an error will be returned.

## FAQ and Troubleshooting

### What if Gee cannot find the gee.toml?
Ensure that the `gee.toml` file exists in the current or parent directories. Use `gee init` to create a new configuration if needed.

### Discovery doesn't show the `d` key
Discovery requires `gh` (GitHub CLI) or `glab` (GitLab CLI). Install one:
```shell
brew install gh    # GitHub
brew install glab  # GitLab
```

### The sub-shell opens but feels like a blank terminal
The `Enter` key runs your `$SHELL` in the repo's directory. Type `exit` or `Ctrl-D` to return to the Gee dashboard.

## License

Gee is licensed under the MIT License. See the LICENSE file for more information.
