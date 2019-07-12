package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	ess "github.com/unixpickle/essentials"
	"github.com/urfave/cli"

	// Providers (service implementations):
	"github.com/stevenxie/api/provider/airtable"
	gh "github.com/stevenxie/api/provider/github"
	gcal "github.com/stevenxie/api/provider/google/calendar"
	gmaps "github.com/stevenxie/api/provider/google/maps"
	"github.com/stevenxie/api/provider/here"
	"github.com/stevenxie/api/provider/rescuetime"
	"github.com/stevenxie/api/provider/spotify"

	"github.com/stevenxie/api/config"
	"github.com/stevenxie/api/internal/info"
	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/cmdutil"
	"github.com/stevenxie/api/pkg/geo"
	"github.com/stevenxie/api/server"
	"github.com/stevenxie/api/stream"
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
		raven    = buildRaven()
		log      = buildLogger(raven)
		cfg, err = config.Load()
	)
	if err != nil {
		return errors.Wrap(err, "loading config")
	}

	// Initialize services:
	log.Info("Initializing services...")

	// Finalizers should be stopped before the program terminates.
	var finalizers []interface{ Stop() }

	// Create availability service.
	var availability api.AvailabilityService
	{
		client, err := gcal.NewClient()
		if err != nil {
			return errors.Wrap(err, "creating GCal client")
		}
		availability = gcal.NewAvailabilityService(
			client,
			cfg.Availability.GCal.CalendarIDs,
		)
	}
	timezone, err := availability.Timezone()
	if err != nil {
		return errors.Wrap(err, "fetching current timezone")
	}

	// Build location service.
	var location api.LocationService
	{
		geocoder, err := here.New(cfg.Location.Here.AppID)
		if err != nil {
			return errors.Wrap(err, "creating MapBox client")
		}
		var source geo.SegmentSource
		if source, err = gmaps.NewHistorian(func(cfg *gmaps.HistorianConfig) {
			cfg.Timezone = timezone
		}); err != nil {
			return errors.Wrap(err, "creating historian")
		}
		if polling := &cfg.Location.Polling; polling.Enabled {
			preloader := stream.NewSegmentsPreloader(
				source, polling.Interval,
				func(cfg *stream.SPConfig) {
					cfg.Logger = log.WithField("service", "segments_preloader").Logger
				},
			)
			source = preloader
			finalizers = append(finalizers, preloader)
		}
		location = geo.NewLocationService(source, geocoder)
	}

	// Build location access service.
	var locationAccess api.LocationAccessService
	{
		client, err := airtable.NewClient()
		if err != nil {
			return errors.Wrap(err, "creating Airtable client")
		}
		config := &cfg.Location.Airtable
		locationAccess = airtable.NewLocationAccessService(
			client,
			config.BaseID,
			config.Table,
			config.View,

			func(cfg *airtable.LASConfig) {
				cfg.Timezone = timezone
				cfg.Logger = log.WithField("service", "location_access").Logger
			},
		)
	}

	// Build GitHub client, a shared dependenncy.
	github, err := gh.New()
	if err != nil {
		return errors.Wrap(err, "creating GitHub client")
	}

	// Build about service.
	var about api.AboutService
	{
		gist := &cfg.About.Gist
		about = gh.NewAboutService(
			github, gist.ID, gist.File,
			location,
		)
	}

	// Build commits service.
	var commits api.GitCommitsService
	{
		commits = github
		if polling := &cfg.Commits.Polling; polling.Enabled {
			preloader := stream.NewCommitsPreloader(
				github,
				polling.Interval,

				func(cfg *stream.CPConfig) {
					cfg.Limit = polling.Limit
					cfg.Logger = log.WithField("service", "commits_preloader").Logger
				},
			)
			commits = preloader
			finalizers = append(finalizers, preloader)
		}
	}

	// Create music service.
	var music api.MusicService
	{
		spotify, err := spotify.New()
		if err != nil {
			return errors.Wrap(err, "creating Spotify client")
		}
		music = spotify
		if polling := &cfg.Music.Polling; polling.Enabled {
			streamer := stream.NewMusicStreamer(
				spotify,
				cfg.Music.Polling.Interval,
				func(cfg *stream.MSConfig) {
					cfg.Logger = log.WithField("service", "music_streamer").Logger
				},
			)
			music = streamer
			finalizers = append(finalizers, streamer)
		}
	}

	// Create productivity service.
	var productivity api.ProductivityService
	{
		productivity, err = rescuetime.New(func(cfg *rescuetime.ClientConfig) {
			cfg.Timezone = timezone
		})
		if err != nil {
			return errors.Wrap(err, "creating RescueTime client")
		}
	}

	// Create and configure server.
	log.Info("Initializing server...")
	srv := server.New(
		about,
		availability,
		commits,
		music,
		productivity,

		location,
		locationAccess,

		func(cfg *server.Config) {
			cfg.Logger = log
			cfg.Raven = raven
		},
	)

	// Shut down server gracefully upon interrupt.
	go shutdownUponInterrupt(srv, log, cfg.ShutdownTimeout)

	err = srv.ListenAndServe(fmt.Sprintf(":%d", c.Int("port")))
	if (err != nil) && (err != http.ErrServerClosed) {
		return errors.Wrap(err, "starting server")
	}

	// Run all finalizers.
	log.Info("Stopping finalizers...")
	for _, finalizer := range finalizers {
		finalizer.Stop()
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
