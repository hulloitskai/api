package main

import (
	"fmt"
	"os"

	errors "golang.org/x/xerrors"

	"github.com/rs/zerolog"
	ess "github.com/unixpickle/essentials"
	"github.com/urfave/cli"

	"github.com/stevenxie/api/config"
	"github.com/stevenxie/api/data/github"
	"github.com/stevenxie/api/data/rescuetime"
	"github.com/stevenxie/api/data/spotify"
	"github.com/stevenxie/api/internal/cmdutil"
	"github.com/stevenxie/api/internal/info"
	"github.com/stevenxie/api/server"
)

func main() {
	// Prepare envvars.
	cmdutil.PrepareEnv()

	app := cli.NewApp()
	app.Name = "apisrv"
	app.Usage = "A personal API server."
	app.UsageText = "apisrv [global options]"
	app.Version = info.Version
	app.Action = run

	// Hide help command.
	app.Commands = []cli.Command{cli.Command{Name: "help", Hidden: true}}

	// Configure flags.
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:   "port",
			Usage:  "port that the server listens on",
			EnvVar: "PORT",
			Value:  3000,
		},
		cli.BoolFlag{
			Name:  "help,h",
			Usage: "show help",
		},
	}

	if err := app.Run(os.Args); err != nil {
		ess.Die("Error:", err)
	}
}

func run(c *cli.Context) error {
	// Init logger, load config.
	var (
		logger   = buildLogger()
		cfg, err = config.Load()
	)
	if err != nil {
		return errors.Errorf("loading config: %w", err)
	}

	// Construct services.
	logger.Info().Msg("Constructing services...")
	github, err := github.New()
	if err != nil {
		return errors.Errorf("creating GitHub client: %w", err)
	}
	infoStore := cfg.BuildInfoStore(github)

	spotifyClient, err := spotify.New()
	if err != nil {
		return errors.Errorf("creating Spotify client: %w", err)
	}

	rtClient, err := rescuetime.New()
	if err != nil {
		return errors.Errorf("creating RescueTime client: %w", err)
	}

	// Create server.
	logger.Info().Msg("Initializing server...")
	var (
		port = c.Int("port")
		srv  = server.New(infoStore, rtClient, spotifyClient, logger)
	)
	// TODO: Shut down server gracefully.
	if err = srv.ListenAndServe(fmt.Sprintf(":%d", port)); err != nil {
		return errors.Errorf("starting server: %w", err)
	}
	return nil
}

// buildLogger builds an application-level zerolog.Logger.
func buildLogger() zerolog.Logger {
	return zerolog.New(zerolog.NewConsoleWriter()).
		With().Timestamp().Logger()
}
