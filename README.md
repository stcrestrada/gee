# Gee: Efficiently Manage Multiple Git Repositories

Gee is a powerful command-line tool designed to help you manage multiple git repositories seamlessly. It allows you to clone, pull, check the status, and run arbitrary commands across repositories listed in a `gee.toml` configuration file, leveraging concurrency for faster operations.

## Features

- **Clone**: Clone multiple repositories listed in the `gee.toml` file.
- **Pull**: Pull changes from the main branch for all repositories.
- **Status**: Compact summary view showing branch, ahead/behind, staged/modified/untracked counts per repo. Detects detached HEAD, rebase, merge, and cherry-pick states.
- **Exec**: Run any shell command across all repos concurrently (e.g., `gee exec git push origin main`).
- **Remove**: Remove repositories from the `gee.toml` configuration.
- **Add**: Add repository to `gee.toml` configuration.

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

## Usage

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

## License

Gee is licensed under the MIT License. See the LICENSE file for more information.
