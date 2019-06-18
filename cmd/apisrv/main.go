package main

import (
	"context"
	"fmt"
	"os"

	errors "golang.org/x/xerrors"

	sentry "github.com/evalphobia/logrus_sentry"
	"github.com/sirupsen/logrus"

	"github.com/getsentry/raven-go"
	ess "github.com/unixpickle/essentials"
	"github.com/urfave/cli"

	"github.com/stevenxie/api/config"
	"github.com/stevenxie/api/data/github"
	"github.com/stevenxie/api/data/rescuetime"
	"github.com/stevenxie/api/data/spotify"
	"github.com/stevenxie/api/internal/cmdutil"
	"github.com/stevenxie/api/internal/info"
	"github.com/stevenxie/api/pkg/gitutil"
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
		ravenClient = buildRaven()
		log         = buildLogger(ravenClient)
		cfg, err    = config.Load()
	)
	if err != nil {
		return errors.Errorf("loading config: %w", err)
	}

	// Construct services:
	log.Info("Constructing services...")

	// Create commits loader and about service.
	githubClient, err := github.New()
	if err != nil {
		return errors.Errorf("creating GitHub client: %w", err)
	}
	// TODO: Use a context that corresponds to the server lifetime.
	commitLoader := cfg.BuildCommitLoader(
		context.Background(),
		githubClient,
		gitutil.WithLogger(log.WithField("service", "commit_loader").Logger),
	)
	aboutService := cfg.BuildAboutService(githubClient)

	// Create Spotify client.
	spotifyClient, err := spotify.New()
	if err != nil {
		return errors.Errorf("creating Spotify client: %w", err)
	}

	// Create GCal client.
	gcalClient, err := cfg.BuildGCalClient()
	if err != nil {
		return errors.Errorf("creating GCal client: %w", err)
	}

	// Create and configure RescueTime client.
	rtClient, err := rescuetime.New()
	if err != nil {
		return errors.Errorf("creating RescueTime client: %w", err)
	}
	timezone, err := gcalClient.Timezone()
	if err != nil {
		return errors.Errorf("failed to load current timezone from GCal: %w", err)
	}
	rtClient.SetTimezone(timezone)

	// Create and configure server.
	log.Info("Initializing server...")
	srv := server.New(
		aboutService,
		rtClient,
		gcalClient,
		commitLoader,
		spotifyClient,
	)
	srv.SetLogger(log)
	srv.UseRaven(ravenClient)

	// TODO: Shut down server gracefully.
	if err = srv.ListenAndServe(fmt.Sprintf(":%d", c.Int("port"))); err != nil {
		return errors.Errorf("starting server: %w", err)
	}
	return nil
}

func buildRaven() *raven.Client {
	rc, err := raven.New(os.Getenv("SENTRY_DSN"))
	if err != nil {
		ess.Die("Failed to build Raven client:", err)
	}

	// Configure client.
	if env := os.Getenv("GOENV"); env != "" {
		rc.SetEnvironment(env)
	}
	rc.SetRelease(info.Version)

	return rc
}

// buildLogger builds an application-level zerolog.Logger, which also captures
// ErrorLevel (and higher) events using Raven.
func buildLogger(rc *raven.Client) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	// Set logger level.
	if os.Getenv("GOENV") == "development" {
		log.SetLevel(logrus.DebugLevel)
	}

	// Add Sentry hook.
	hook, err := sentry.NewAsyncWithClientSentryHook(rc,
		[]logrus.Level{logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel})
	if err != nil {
		ess.Die("Failed to build Sentry Logrus hook:", err)
	}
	log.AddHook(hook)

	return log
}
