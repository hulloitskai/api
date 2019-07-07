package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/stevenxie/api/pkg/geo"

	"github.com/stevenxie/api/provider/airtable"

	errors "golang.org/x/xerrors"

	"github.com/sirupsen/logrus"
	ess "github.com/unixpickle/essentials"
	"github.com/urfave/cli"

	"github.com/stevenxie/api/config"
	"github.com/stevenxie/api/internal/info"
	"github.com/stevenxie/api/pkg/cmdutil"
	gh "github.com/stevenxie/api/provider/github"
	gcal "github.com/stevenxie/api/provider/google/calendar"
	gmaps "github.com/stevenxie/api/provider/google/maps"
	"github.com/stevenxie/api/provider/here"
	"github.com/stevenxie/api/provider/rescuetime"
	"github.com/stevenxie/api/provider/spotify"
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

	// Build location service.
	geocoder, err := here.New(cfg.Location.Here.AppID)
	if err != nil {
		return errors.Errorf("creating MapBox client: %w", err)
	}
	historian, err := gmaps.NewHistorian()
	if err != nil {
		return errors.Errorf("creating historian: %w", err)
	}
	location := geo.NewLocationService(historian, geocoder)

	// Build location access service.
	airc, err := airtable.NewClient()
	if err != nil {
		return errors.Errorf("creating Airtable client: %w", err)
	}
	var (
		airtableCfg    = &cfg.Location.Airtable
		locationAccess = airtable.NewLocationAccessService(
			airc,
			airtableCfg.BaseID,
			airtableCfg.Table,
			airtableCfg.View,
		)
	)

	// Build about service.
	github, err := gh.New()
	if err != nil {
		return errors.Errorf("creating GitHub client: %w", err)
	}
	gistID, gistFile := cfg.AboutGistInfo()
	about := gh.NewAboutService(
		github, gistID, gistFile,
		location,
	)

	// Build commits service.
	commits := stream.NewCommitsPreloader(
		github,
		cfg.Commits.PollInterval,
		stream.WithCPLimit(cfg.Commits.Limit),
		stream.WithCPLogger(log.WithField("service", "commits_preloader").Logger),
	)

	// Create now-playing service.
	spotify, err := spotify.New()
	if err != nil {
		return errors.Errorf("creating Spotify client: %w", err)
	}
	nowplaying := stream.NewNowPlayingStreamer(
		spotify,
		cfg.Music.PollInterval,
		stream.WithNPSLogger(
			log.WithField("service", "nowplaying_streamer").Logger,
		),
	)

	// Create gcal client.
	gcalc, err := gcal.NewClient()
	if err != nil {
		return errors.Errorf("creating GCal client: %w", err)
	}
	availability := gcal.NewAvailabilityService(
		gcalc,
		cfg.GCalCalendarIDs(),
	)

	// Create and configure RescueTime client.
	timezone, err := availability.Timezone()
	if err != nil {
		return errors.Errorf("failed to load current timezone from GCal: %w", err)
	}
	rescuetime, err := rescuetime.New(rescuetime.WithTimezone(timezone))
	if err != nil {
		return errors.Errorf("creating RescueTime client: %w", err)
	}

	// Create and configure server.
	log.Info("Initializing server...")
	srv := server.New(
		about,
		availability,
		commits,
		nowplaying,
		rescuetime,

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

	// Stop preloaders and streamers.
	commits.Stop()
	nowplaying.Stop()

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
