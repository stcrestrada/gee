package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/go-playground/validator/v10"
	toml "github.com/pelletier/go-toml"
	"github.com/stcrestrada/gogo"
	"github.com/urfave/cli/v2"
)

var (
	app *cli.App
)
var validate *validator.Validate

type CommandOutput struct {
	Repo   string
	Dir    string
	Output []byte
}

type GitCommand struct {
	Repo string
	Dir  string
}

type Repo struct {
	// name of repo
	Name string `toml:"name" validate:"required,min=1"`
	// path of repo
	Path string `toml:"path" validate:"required,min=1"`
}

type Config struct {
	Repos []Repo `toml:"repos" validate:"required,dive,required"`
}

func (c *GitCommand) PullAll() ([]byte, error) {
	cmd := exec.Command("git", "pull")
	cmd.Dir = c.Dir
	output, err := cmd.Output()

	return output, err
}

func (c *GitCommand) CurrentBranch() ([]byte, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = c.Dir
	output, err := cmd.Output()

	return output, err
}

func (c *GitCommand) MainBranch() ([]byte, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = c.Dir
	output, err := cmd.Output()

	return output, err
}

func (c *GitCommand) Status() ([]byte, error) {
	cmd := exec.Command("git", "status")
	cmd.Dir = c.Dir
	output, err := cmd.Output()

	return output, err
}

func GeeStatusAll(repo Repo) (*CommandOutput, error) {
	cmd := GitCommand{
		Repo: repo.Name,
		Dir:  repo.Path,
	}
	statusOutput, err := cmd.Status()
	if err != nil {
		CheckIfError(err)
		return nil, err
	}
	return &CommandOutput{
		Repo:   repo.Name,
		Dir:    repo.Path,
		Output: statusOutput,
	}, nil
}

func GeePullAll(repo Repo) (*CommandOutput, error) {
	cmd := GitCommand{
		Repo: repo.Name,
		Dir:  repo.Path,
	}
	mainBranch, err := cmd.MainBranch()
	if err != nil {
		CheckIfError(err)
		return nil, err
	}

	currentBranch, err := cmd.CurrentBranch()
	if err != nil {
		CheckIfError(err)
		return nil, err
	}

	if string(mainBranch) != string(currentBranch) {
		err = fmt.Errorf("skipping, cannot update repo, %s must checkout to %s", repo.Name, mainBranch)
		return nil, err
	}

	pullOutput, err := cmd.PullAll()
	if err != nil {
		CheckIfError(err)
		return nil, err
	}
	return &CommandOutput{
		Repo:   repo.Name,
		Dir:    repo.Path,
		Output: pullOutput,
	}, nil
}

func loadConfig() (*Config, error) {
	conf := Config{}
	config, err := toml.LoadFile("gee.toml")
	if err != nil {
		return nil, err
	}

	err = config.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}

	// validate that config has necessary fields
	err = validate.Struct(&conf)
	if err != nil {
		return nil, err
	}

	return &conf, err
}

func main() {
	config, err := loadConfig()
	if err != nil {
		CheckIfError(err)
		return
	}
	app.Commands = []*cli.Command{
		{
			Name:  "pull",
			Usage: "Git pull and update all repos",
			Action: func(c *cli.Context) error {
				concurrency := 3
				repos := config.Repos
				pool := gogo.NewPool(concurrency, len(repos), func(i int) func() (interface{}, error) {
					repo := repos[i]
					return func() (interface{}, error) {
						output, err := GeePullAll(repo)
						return output, err
					}
				})
				feed := pool.Go()
				for res := range feed {
					if res.Error == nil {
						cmdOutput := res.Result.(*CommandOutput)
						Info("Pulling Repo %s \n", cmdOutput.Repo)
						println(string(cmdOutput.Output))
						Info("Finished pulling %s \n", cmdOutput.Repo)
						continue
					}
					Warning(res.Error.Error())
				}
				return nil
			},
		},
		{
			Name:  "status",
			Usage: "Git status of all repos",
			Action: func(c *cli.Context) error {
				concurrency := 2
				repos := config.Repos
				pool := gogo.NewPool(concurrency, len(repos), func(i int) func() (interface{}, error) {
					repo := repos[i]
					return func() (interface{}, error) {
						output, err := GeeStatusAll(repo)
						return output, err
					}
				})
				feed := pool.Go()
				for res := range feed {
					if res.Error == nil {
						cmdOutput := res.Result.(*CommandOutput)
						Info("Status of %s \n", cmdOutput.Repo)
						println(string(cmdOutput.Output))
						continue
					}
					Warning(res.Error.Error())
				}
				return nil
			},
		},
	}

	// Run the CLI app
	err = app.Run(os.Args)
	if err != nil {
		CheckIfError(err)
		return
	}

}

func init() {
	validate = validator.New()
	// Initialise a CLI app
	app = cli.NewApp()
	app.Name = "gee"
	app.Usage = "Gee gives control of git across repos without changing directories"
	app.Version = "0.0.0"
}

func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}
