package config

import (
	"fmt"
	"reflect"

	"github.com/arkan/mcli"
	"github.com/urfave/cli/v2"
)

func newListCommand(app *mcli.App) *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Print the list of configuration keys and values",
		Action: func(c *cli.Context) error {
			if c.NArg() != 0 {
				return fmt.Errorf("usage: lsc config list")
			}

			e := reflect.ValueOf(app.Config)
			for i := 0; i < e.NumField(); i++ {
				f := e.Field(i)
				fmt.Printf("%s: %s\n", e.Type().Field(i).Name, f.Interface())
			}
			return nil
		},
	}
}
