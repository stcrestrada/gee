# Gee: Efficiently Manage Multiple Git Repositories

Gee is a powerful tool for managing multiple git repositories. It features a full-screen interactive TUI (Terminal User Interface) with live status updates, Vim-style navigation, and one-key actions — plus a traditional CLI for scripting and automation. All operations run concurrently using a worker pool for speed.

<!-- TODO: Screenshot — full dashboard view showing several repos with mixed status (clean, modified, ahead/behind) -->

## Features

- **Interactive Dashboard**: Run `gee` to launch a K9s-style full-screen TUI with live-updating repo status
- **Automatic Discovery**: Gee scans your home directory for git repos and streams them into the dashboard as they're found
- **Pinned Repos**: Pin your important repos with `gee add` so they always show up first in the dashboard and CLI commands
- **Vim Navigation**: `j`/`k` to move, `g`/`G` to jump, `/` to filter repos by name
- **Teleport**: Press `Enter` on any repo to instantly `cd` into it (requires shell integration)
- **One-Key Actions**: `p` to pull, `e` to exec a command
- **Remote Discovery**: Press `d` to browse your GitHub/GitLab repos, multi-select, and batch-clone them
- **Staleness Detection**: Repos with dirty changes and no recent activity are flagged as `STALE`
- **Context-Aware CLI**: Run `gee status` inside a repo to target just that repo, or use `--all` for everything
- **Zero Config**: No config files to maintain — Gee uses a JSON cache at `~/.config/gee/cache.json`
- **Concurrent Operations**: Every operation runs in parallel across all repos

## Installation

### Homebrew (macOS / Linux)

```
brew tap stcrestrada/gee
brew install gee
```

Upgrade:
```
brew update && brew upgrade gee
```

### Go Install (any platform)

```
go install github.com/stcrestrada/gee@latest
```

Requires Go 1.21+.

### Download Binary

Grab the latest release for your platform from the [Releases page](https://github.com/stcrestrada/gee/releases). Extract and place the `gee` binary somewhere on your `PATH`.

Available for: **macOS** (amd64, arm64), **Linux** (amd64, arm64), **Windows** (amd64).

### Uninstall
```
brew uninstall gee
```

### Shell Integration

Add this to your shell config to enable **teleport** (`Enter` to `cd` into a repo):

**Bash** (`~/.bashrc`) / **Zsh** (`~/.zshrc`):
```sh
eval "$(gee --init)"
```

**Fish** (`~/.config/fish/config.fish`):
```fish
gee --init | source
```

**PowerShell** (`$PROFILE`):
```powershell
function gee {
  if ($args.Count -eq 0) {
    $env:GEE_TELEPORT = "1"
    $result = & gee.exe
    $env:GEE_TELEPORT = $null
    if ($result -and (Test-Path $result)) { Set-Location $result }
  } else {
    & gee.exe @args
  }
}
```

Without shell integration, the TUI will print the path but won't change your directory.

## Quick Start

```shell
# Pin repos you care about
cd path/to/repo && gee add

# Launch the interactive dashboard
gee
```

Gee automatically discovers git repos under your home directory. Use `gee add` inside any repo to pin it — pinned repos appear first in the dashboard and are the default target for CLI commands.

## Interactive TUI

Run `gee` with no arguments to launch the interactive dashboard.

<!-- TODO: Screenshot — dashboard showing the header, repo table with branch/sync/changes columns, and help bar at the bottom -->

### Dashboard

The dashboard shows all your repos in a live-updating table with:
- Pin indicator (`*`) for pinned repos
- Branch name (with rebase/merge/cherry-pick state detection)
- Sync status (ahead/behind remote)
- Change counts (staged, modified, untracked, conflicts)
- `STALE` badge for repos with dirty changes and no recent file activity

The header shows total repo count, pinned count, and a scanning indicator while discovery is in progress. Status refreshes automatically every 5 seconds and after every action.

### Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` | Move cursor down / up |
| `g` / `G` | Jump to first / last repo |
| `a` | Toggle pin on the selected repo |
| `p` | Pull the selected repo |
| `P` | Pull all visible repos |
| `e` | Open exec prompt — run any shell command in the selected repo |
| `Enter` | Teleport — quit TUI and `cd` into the selected repo |
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
| `Enter` | Clone all selected repos and pin them |
| `Esc` | Return to the dashboard |

Discovery requires `gh` (GitHub CLI) or `glab` (GitLab CLI) to be installed. If neither is available, the `d` key is hidden from the help bar. GitHub is preferred when both are present.

## CLI Commands

All commands also work as traditional CLI subcommands for scripting and CI pipelines. Commands are **context-aware**: if you run them inside a cached repo, they target that repo. Otherwise, they target all pinned repos. Use `--all` to target every cached repo.

### Pin a Repository
Pin the current git repo so it shows up in the dashboard and CLI commands:
```
cd path/to/repo
gee add
```

Pin all git repos in the current directory:
```
gee add --all
```

Interactively select which repos to pin:
```
gee add --all-select
```

### Check Status
Show a compact summary of your repos:
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

Target all cached repos (not just pinned):
```
gee status --all
```

### Pull Changes
Pull changes for your repos:
```
gee pull
gee pull --all
```

### Execute Commands
Run any command across repos concurrently:
```shell
gee exec git push origin main
gee exec git stash
gee exec npm install
gee exec "git stash && git pull && git stash pop"
gee exec --all git fetch
```

### Unpin a Repository
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

## How It Works

Gee stores all repo data in a JSON cache at `~/.config/gee/cache.json`. There are two types of repos:

- **Pinned**: Repos you've explicitly added with `gee add`. These are the default target for CLI commands and always appear first in the dashboard.
- **Discovered**: Repos found automatically by scanning your filesystem. These appear in the dashboard but are not targeted by CLI commands unless you use `--all`.

### Migration from gee.toml

If you're upgrading from an older version of Gee that used `gee.toml`, the first time you launch the TUI it will automatically import your repos from `gee.toml` into the cache as pinned repos. No manual migration is needed.

## FAQ and Troubleshooting

### Discovery doesn't show the `d` key
Discovery requires `gh` (GitHub CLI) or `glab` (GitLab CLI). Install one:
```shell
brew install gh    # GitHub
brew install glab  # GitLab
```

### Where is the cache stored?
`~/.config/gee/cache.json`. You can inspect or edit it directly — it's plain JSON.

## License

Gee is licensed under the MIT License. See the LICENSE file for more information.
