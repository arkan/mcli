package mcli

import (
	"context"
	"os"

	"github.com/arkan/dotconfig"
)

const (
	appName        = "mcli"
	DefaultVersion = "0.0.0"
)

type Config struct {
	// PersonalGithubToken stores the personal user token.
	PersonalGithubToken string `yaml:"personal_github_token"`
}

func (c *Config) GetGithubToken() string {
	if v, ok := os.LookupEnv("GH_TOKEN"); ok {
		return v
	}
	return c.PersonalGithubToken
}

type App struct {
	Version string
	Config  Config
	Context context.Context
}

func New(version string) (*App, error) {
	app := &App{
		Version: version,
	}

	if err := dotconfig.Load(appName, &app.Config); err != nil {
		if err == dotconfig.ErrConfigNotFound {
			// Set default values
			if err := dotconfig.Save(appName, &app.Config); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return app, nil
}

func (a *App) IsConfigured() (bool, error) {
	t := a.Config.GetGithubToken()

	return t != "", nil
}

func (a *App) IsAutoUpdateEnabled() bool {
	if os.Getenv("AUTOUPDATE") == "0" {
		return false
	}

	return a.IsRelease()
}

func (a *App) IsRelease() bool {
	return a.Version != DefaultVersion
}

func (a *App) SaveConfig() error {
	if err := dotconfig.Save(appName, &a.Config); err != nil {
		return err
	}

	return nil
}
