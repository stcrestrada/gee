# Gee: Efficiently Manage Multiple Git Repositories

Gee is a powerful command-line tool designed to help you manage multiple git repositories seamlessly. It allows you to clone, pull, and check the status of repositories listed in a `gee.toml` configuration file, leveraging concurrency for faster operations.

## Features

- **Clone**: Clone multiple repositories listed in the `gee.toml` file.
- **Pull**: Pull changes from the main branch for all repositories.
- **Status**: Check the git status of all repositories.
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
```brew uninstall gee```

### Initialize Configuration 
#### Create an initial gee.toml configuration file:

```
gee init
```

### Add Repository 
#### Add a repository to the `gee.toml` file. Note that the `gee add` command must be run from inside a `.git` directory:
```
cd path/to/repo
gee add
```

### Check Status 
#### Check the status of all git directories added to gee.toml
```
gee status
```

### Pull Changes 
##### Pull changes from the main branch for all repositories:
```
gee pull
```

### Remove Repository 
##### Automatically detect from the current directory:
```
gee remove 
```
##### Specify repository name using flag:
```shell
gee remove -r repo_name
```
##### OR
```shell
gee remove --repo repo_name
```


### Clone Repositories
##### Clone all repositories listed in your gee.toml file:
```shell
gee clone
```

## Configuration
#### Manually Configure gee.toml
```
[[repos]]
name = "repo_name"
path = "/path/to/repo"
remote = "https://github.com/user/repo.git"
````

### Finding gee.toml
#### Gee searches for gee.toml in the current and parent directories. If no configuration file is found, an error will be returned.

## FAQ and Troubleshooting
### What if Gee cannot find the gee.toml?
#### Ensure that the gee.toml file exists in the current or parent directories. Use gee init to create a new configuration if needed.

## License
#### Gee is licensed under the MIT License. See the LICENSE file for more information.

