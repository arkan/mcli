package config

import (
	"github.com/arkan/mcli"
	"github.com/urfave/cli/v2"
)

// NewCommand returns the Auth root command.
func NewCommand(app *mcli.App) *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Display or set configuration values",
		Subcommands: []*cli.Command{
			newListCommand(app),
		},
	}
}
