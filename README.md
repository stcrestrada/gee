# gee Manage your repos better

### install

```
brew tap stcrestrada/gee https://github.com/stcrestrada/gee

brew install gee
```

## initialize 
#### Creates a gee.toml file with all the git directories
```
[[repos]]
  name = "gee"
  path = "/Users/stephenestrada/projects/gee"
```
```
gee init
```

## Add repos 
#### cd into the git directory and run this command
```
gee add
```

## Git status
#### Check the status of all git directories added to gee.toml
```
gee status
```

## Git pull
```
gee pull
```
