package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	errors "golang.org/x/xerrors"

	"github.com/sirupsen/logrus"
	ess "github.com/unixpickle/essentials"
	"github.com/urfave/cli"

	"github.com/stevenxie/api/config"
	"github.com/stevenxie/api/internal/info"
	"github.com/stevenxie/api/pkg/cmdutil"
	"github.com/stevenxie/api/provider/gcal"
	"github.com/stevenxie/api/provider/github"
	"github.com/stevenxie/api/provider/rescuetime"
	"github.com/stevenxie/api/provider/spotify"
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

	// Initialize services:
	log.Info("Initializing services...")

	githubClient, err := github.New()
	if err != nil {
		return errors.Errorf("creating GitHub client: %w", err)
	}

	// Build about service.
	gistID, gistFile := cfg.AboutGistInfo()
	aboutService := github.NewAboutService(githubClient, gistID, gistFile)

	// Create Spotify client.
	spotifyClient, err := spotify.New()
	if err != nil {
		return errors.Errorf("creating Spotify client: %w", err)
	}

	// Create GCal client.
	gcalClient, err := gcal.NewClient()
	if err != nil {
		return errors.Errorf("creating GCal client: %w", err)
	}
	availabilityService := gcal.NewAvailabilityService(
		gcalClient,
		cfg.GCalCalendarIDs(),
	)

	// Create and configure RescueTime client.
	timezone, err := availabilityService.Timezone()
	if err != nil {
		return errors.Errorf("failed to load current timezone from GCal: %w", err)
	}
	rtClient, err := rescuetime.New(rescuetime.WithTimezone(timezone))
	if err != nil {
		return errors.Errorf("creating RescueTime client: %w", err)
	}

	// Create and configure server.
	log.Info("Initializing server...")
	srv := server.New(
		aboutService,
		rtClient,
		availabilityService,
		githubClient,
		spotifyClient,
		append(
			cfg.ServerOpts(),
			server.WithLogger(log),
			server.WithRaven(ravenClient),
		)...,
	)

	// Shut down server gracefully upon interrupt.
	go shutdownUponInterrupt(srv, log, cfg.ShutdownTimeout)

	// TODO: Shut down server gracefully.
	err = srv.ListenAndServe(fmt.Sprintf(":%d", c.Int("port")))
	if (err != nil) && (err != http.ErrServerClosed) {
		return errors.Errorf("starting server: %w", err)
	}

	return nil
}

func shutdownUponInterrupt(
	srv *server.Server,
	log *logrus.Logger,
	timeout *time.Duration,
) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// Wait for interrupt signal.
	<-sig

	const msg = "Received interrupt signal; shutting down."
	if timeout != nil {
		log.WithField("timeout", timeout.String()).Info(msg)
	} else {
		log.Info(msg)
	}

	// Prepare shutdown context.
	ctx := context.Background()
	if timeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), *timeout)
		defer cancel()
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Server didn't shut down correctly.")
	}
}
