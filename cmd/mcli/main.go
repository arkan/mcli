package main

import (
	"errors"
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/arkan/mcli"
	"github.com/arkan/mcli/internal/commands"
	"github.com/arkan/mcli/internal/github"

	"golang.org/x/exp/slog"
)

var (
	Version     string = mcli.DefaultVersion
	githubOwner string = "arkan"
	githubRepo  string = "mcli"
	binaryName  string = "mcli"
)

func main() {
	app, err := mcli.New(Version)
	if err != nil {
		slog.Error("Unable to start mcli", "error", err)
		os.Exit(1)
	}

	isConfigured, err := app.IsConfigured()
	if err != nil {
		slog.Error("Unable to get configuration status", "error", err)
		os.Exit(1)
	}

	if !isConfigured {
		slog.Info("Not configured. Let's do the configuration!")

		msg := heredoc.Docf(`
			You need to generate a new Token for %s.
			Please go to https://github.com/settings/tokens/new?scopes=repo,user,write:packages&description=%s
			And create a Github personal access token.
		`, binaryName, binaryName)
		prompt := &survey.Input{
			Message: msg,
		}

		fn := func(val interface{}) error {
			token := val.(string)
			if !github.New(token, githubOwner, githubRepo).IsTokenValid() {
				return errors.New("token invalid or expired")
			}
			return nil
		}

		err := survey.AskOne(prompt, &app.Config.PersonalGithubToken, survey.WithValidator(survey.Required), survey.WithValidator(fn))
		if err != nil {
			slog.Error("Unable to get user input", "error", err)
			os.Exit(1)
		}
		if err := app.SaveConfig(); err != nil {
			slog.Error("Unable to save config", "error", err)
			os.Exit(1)
		}
	}

	if app.IsAutoUpdateEnabled() {
		g := github.New(app.Config.GetGithubToken(), githubOwner, githubRepo)
		err := g.DownloadNewVersionIfNeeded(binaryName, Version)
		if err != nil {
			slog.Error("Unable to upgrade", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("Starting now!")

	c := commands.New(app)
	if err := c.Run(os.Args); err != nil {
		log.Fatalf("Error: %v", err.Error())
	}
}
