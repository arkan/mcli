package commands

import (
	"fmt"
	"os"

	"github.com/arkan/mcli"
	"github.com/arkan/mcli/internal/commands/config"
	"github.com/urfave/cli/v2"
)

type Cmd struct {
	cli *cli.App
	app *mcli.App
}

// New takes an *lsc.App instance as parameter and returns an instance of Cmd.
func New(app *mcli.App) *Cmd {
	cliApp := &cli.App{
		Name:                 "mcli",
		Usage:                "The Command Line Interface built for {{YOUR COMPANY}}",
		Version:              app.Version,
		EnableBashCompletion: true,
		Before: func(c *cli.Context) error {
			c.Context = app.Context
			return nil
		},
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("Version: %v\n", c.App.Version)
	}

	cliApp.Commands = []*cli.Command{
		config.NewCommand(app),
	}

	cliApp.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Printf(
			"%s: '%s' is not a %s command. See '%s --help'.\n",
			c.App.Name,
			command,
			c.App.Name,
			os.Args[0],
		)
		os.Exit(1)
	}

	return &Cmd{cli: cliApp, app: app}
}

// Run runs the cli app.
func (c *Cmd) Run(args []string) error {
	return c.cli.Run(args)
}
