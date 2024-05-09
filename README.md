# gee Manage your repos better

### install

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

### initialize 
#### Creates a gee.toml file with all the git directories added to gee.toml
###### Run this command in the directory you want to manage with gee in
```
gee init
```

## Add repos 
#### cd into the git directory and run this command, will add repos to gee.toml
```
gee add
```

## Git status
#### Check the status of all git directories added to gee.toml
```
gee status
```

## Git pull
##### Pretty cool command. gee pull will pull all git changes as long as you're in the main branch. If you have uncommitted changes in the main branch, gee stashes those changes, pulls down and reapplies the uncommitted changes. 
```
gee pull
```

## Git Clone
###### Requires user to manually add remote to gee.toml 
```
gee clone
```

#### Example. Please manually add `remote` to gee.toml for gee clone to work.
```
[[repos]]
name = "gee"
path = "/Users/stcrestrada/Projects"
remote = "git@github.com:stcrestrada/gee.git"
````

#### Things to consider:
- When manually configuring `gee.toml` make sure that name is the same as the repository name.
- When manually configuring `gee.toml` make sure that path does not include the repository name.
- When manually configuring `gee.toml` make sure that remote to leverage's `gee's` full potential.
- A `gee.toml` can exist in multiple directories. For example if you run `gee create` inside of `/one/project/dev/`, `gee` will only look for `gee.toml` inside `/one/project/dev/`. If `gee command` is run inside `/one/project/` it will look for `gee.toml` inside `/one/project/`, if one does not exist, `gee` will move up to the next parent directory, `/one` and so on until it finds a `gee.toml` file. If no `gee.toml` is found, `gee` will return an error. 
