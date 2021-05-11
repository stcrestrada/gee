# gee Manage your repos better

### install

```
brew tap stcrestrada/gee https://github.com/stcrestrada/gee

brew install gee
```

## initialize 
#### Creates a gee.toml file with all the git directories
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
