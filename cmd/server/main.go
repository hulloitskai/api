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

	"go.stevenxie.me/api/config"
	"go.stevenxie.me/api/internal/info"
	"go.stevenxie.me/api/pkg/cmdutil"
	gh "go.stevenxie.me/api/pkg/github"
	"go.stevenxie.me/api/pkg/google"
	"go.stevenxie.me/api/server"

	"go.stevenxie.me/api/service/about"
	aboutgh "go.stevenxie.me/api/service/about/github"
	"go.stevenxie.me/api/service/availability"
	"go.stevenxie.me/api/service/availability/gcal"
	cm "go.stevenxie.me/api/service/commits"
	commitsgh "go.stevenxie.me/api/service/commits/github"
	"go.stevenxie.me/api/service/music"
	"go.stevenxie.me/api/service/music/spotify"
	"go.stevenxie.me/api/service/productivity"
	"go.stevenxie.me/api/service/productivity/rescuetime"

	loc "go.stevenxie.me/api/service/location"
	"go.stevenxie.me/api/service/location/airtable"
	"go.stevenxie.me/api/service/location/geocode"
	"go.stevenxie.me/api/service/location/geocode/here"
	"go.stevenxie.me/api/service/location/gmaps"
)

func main() {
	// Prepare envvars.
	cmdutil.PrepareEnv()

	app := cli.NewApp()
	app.Name = "server"
	app.Usage = "A personal API server."
	app.UsageText = "server [global options]"
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

func run(c *cli.Context) (err error) {
	// Init logger, load config.
	var (
		raven = buildRaven()
		log   = buildLogger(raven)
	)
	cfg, err := config.Load()
	if err != nil {
		return errors.Wrap(err, "loading config")
	}

	// Initialize services:
	log.Info("Initializing services...")

	// Finalizers should be run before the program terminates.
	var finalizers cmdutil.Finalizers
	defer func() {
		if len(finalizers) == 0 {
			return
		}

		// Run finalizers in reverse order.
		log.Info("Running finalizers...")
		for i := len(finalizers) - 1; i >= 0; i-- {
			if ferr := finalizers[i](); ferr != nil {
				log.WithError(ferr).Error("A finalizer failed.")
				if err == nil {
					err = errors.New("one or more finalizers failed")
				}
			}
		}
	}()

	// Create availability service.
	var availability availability.Service
	{
		clientset, err := google.NewClientSet()
		if err != nil {
			return errors.Wrap(err, "creating Google clientset")
		}
		svc, err := clientset.CalendarService(context.Background())
		if err != nil {
			return errors.Wrap(err, "creating calendar service")
		}
		availability = gcal.NewAvailabilityService(
			svc,
			cfg.Availability.GCal.CalendarIDs,
		)
	}
	timezone, err := availability.Timezone()
	if err != nil {
		return errors.Wrap(err, "fetching current timezone")
	}

	// Build location service.
	var location loc.Service
	{
		hc, err := here.NewClient(cfg.Location.Here.AppID)
		if err != nil {
			return errors.Wrap(err, "creating MapBox hc")
		}
		geocoder := here.NewGeocoder(hc)

		var history loc.HistoryService
		if history, err = gmaps.NewHistorian(func(cfg *gmaps.HistorianConfig) {
			cfg.Timezone = timezone
		}); err != nil {
			return errors.Wrap(err, "creating historian")
		}
		if polling := &cfg.Location.Polling; polling.Enabled {
			preloader := loc.NewHistoryPreloader(
				history, polling.Interval,
				func(hpc *loc.HistoryPreloaderConfig) {
					hpc.Logger = log.WithField(
						"component",
						"location.HistoryPreloader",
					)
				},
			)
			history = preloader
			finalizers = append(finalizers, func() error {
				preloader.Stop()
				return nil
			})
		}

		// Decode geocode level from config string.
		var regionGeocodeLevel geocode.Level
		if level := cfg.Location.Region.GeocodeLevel; level != "" {
			if regionGeocodeLevel, err = geocode.ParseLevel(level); err != nil {
				return errors.Wrapf(err, "parsing region geocode level '%s'", level)
			}
		}

		// Create location service.
		location = geocode.NewLocationService(
			history,
			geocoder,
			func(lsc *geocode.LocationServiceConfig) {
				if regionGeocodeLevel != 0 {
					lsc.RegionGeocodeLevel = regionGeocodeLevel
				}
			},
		)
	}

	// Build location access service.
	var locationAccess loc.AccessService
	{
		airc, err := airtable.NewClient()
		if err != nil {
			return errors.Wrap(err, "creating Airtable hc")
		}

		cfg := &cfg.Location.Airtable
		locationAccess = airtable.NewLocationAccessService(
			airc,
			cfg.BaseID,
			cfg.Table,
			cfg.View,

			func(lasc *airtable.LocationAccessServiceConfig) {
				lasc.Timezone = timezone
				lasc.Logger = log.WithField(
					"component",
					"airtable.LocationAccessService",
				)
			},
		)
	}

	// Build GitHub client, a shared dependenncy.
	github, err := gh.New()
	if err != nil {
		return errors.Wrap(err, "creating GitHub hc")
	}

	// Build about service.
	var about about.Service
	{
		gist := &cfg.About.Gist
		about = aboutgh.NewAboutService(
			github.GHClient().Gists,
			gist.ID, gist.File,
			location,
		)
	}

	// Build commits service.
	var commits cm.Service
	{
		commits = commitsgh.NewCommitsService(github)
		if polling := &cfg.Commits.Polling; polling.Enabled {
			preloader := cm.NewPreloader(
				commits,
				polling.Interval,

				func(pc *cm.PreloaderConfig) {
					pc.Limit = polling.Limit
					pc.Logger = log.WithField(
						"component",
						"commits.Preloader",
					)
				},
			)
			commits = preloader
			finalizers = append(finalizers, func() error {
				preloader.Stop()
				return nil
			})
		}
	}

	// Create nowplaying service.
	var nowplaying music.NowPlayingService
	{
		client, err := spotify.New()
		if err != nil {
			return errors.Wrap(err, "creating Spotify hc")
		}
		nowplaying = spotify.NewNowPlayingService(client)

		if polling := &cfg.Music.Polling; polling.Enabled {
			streamer := music.NewNowPlayingStreamer(
				nowplaying,
				cfg.Music.Polling.Interval,
				func(npsc *music.NowPlayingStreamerConfig) {
					npsc.Logger = log.WithField(
						"component",
						"music.NowPlayingStreamer",
					)
				},
			)
			nowplaying = streamer
			finalizers = append(finalizers, func() error {
				streamer.Stop()
				return nil
			})
		}
	}

	// Create productivity service.
	var productivity productivity.Service
	{
		client, err := rescuetime.NewClient()
		if err != nil {
			return errors.Wrap(err, "creating RescueTime client")
		}
		productivity = rescuetime.NewProductivityService(
			client,
			func(psc *rescuetime.ProductivityServiceConfig) {
				psc.Timezone = timezone
			},
		)
	}

	// Create and configure server.
	log.Info("Initializing server...")
	srv := server.New(
		about,
		productivity,
		availability,
		commits,
		nowplaying,

		location,
		locationAccess,

		func(cfg *server.Config) {
			cfg.Logger = log
			cfg.Raven = raven
		},
	)

	// Shut down server gracefully upon interrupt.
	go shutdownUponInterrupt(srv, log, cfg.Server.ShutdownTimeout)

	err = srv.ListenAndServe(fmt.Sprintf(":%d", c.Int("port")))
	if err == http.ErrServerClosed {
		err = nil
	}
	return errors.Wrap(err, "starting server")
}

func shutdownUponInterrupt(
	srv *server.Server,
	log *logrus.Logger,
	timeout *time.Duration,
) {
	sig := make(chan os.Signal)
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
