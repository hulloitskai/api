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
		return errors.Errorf("loading config: %w", err)
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
			return errors.Errorf("creating GCal client: %w", err)
		}
		availability = gcal.NewAvailabilityService(
			client,
			cfg.Availability.GCal.CalendarIDs,
		)
	}
	timezone, err := availability.Timezone()
	if err != nil {
		return errors.Errorf("fetching current timezone: %w", err)
	}

	// Build location service.
	var location api.LocationService
	{
		geocoder, err := here.New(cfg.Location.Here.AppID)
		if err != nil {
			return errors.Errorf("creating MapBox client: %w", err)
		}
		var source geo.SegmentSource
		source, err = gmaps.NewHistorian(gmaps.WithHTimezone(timezone))
		if err != nil {
			return errors.Errorf("creating historian: %w", err)
		}
		if polling := &cfg.Location.Polling; polling.Enabled {
			preloader := stream.NewSegmentsPreloader(
				source, polling.Interval,
				stream.WithSPLogger(
					log.WithField("service", "segments_preloader").Logger,
				),
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
			return errors.Errorf("creating Airtable client: %w", err)
		}
		config := &cfg.Location.Airtable
		locationAccess = airtable.NewLocationAccessService(
			client,
			config.BaseID,
			config.Table,
			config.View,

			airtable.WithLASTimezone(timezone),
			airtable.WithLASLogger(
				log.WithField("service", "location_access").Logger,
			),
		)
	}

	// Build GitHub client, a shared dependenncy.
	github, err := gh.New()
	if err != nil {
		return errors.Errorf("creating GitHub client: %w", err)
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

				stream.WithCPLimit(polling.Limit),
				stream.WithCPLogger(
					log.WithField("service", "commits_preloader").Logger,
				),
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
			return errors.Errorf("creating Spotify client: %w", err)
		}
		music = spotify
		if polling := &cfg.Music.Polling; polling.Enabled {
			streamer := stream.NewMusicStreamer(
				spotify,
				cfg.Music.Polling.Interval,
				stream.WithMSLogger(log.WithField("service", "music_streamer").Logger),
			)
			music = streamer
			finalizers = append(finalizers, streamer)
		}
	}

	// Create productivity service.
	var productivity api.ProductivityService
	{
		productivity, err = rescuetime.New(rescuetime.WithTimezone(timezone))
		if err != nil {
			return errors.Errorf("creating RescueTime client: %w", err)
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

		server.WithLogger(log),
		server.WithRaven(raven),
	)

	// Shut down server gracefully upon interrupt.
	go shutdownUponInterrupt(srv, log, cfg.ShutdownTimeout)

	err = srv.ListenAndServe(fmt.Sprintf(":%d", c.Int("port")))
	if (err != nil) && (err != http.ErrServerClosed) {
		return errors.Errorf("starting server: %w", err)
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
