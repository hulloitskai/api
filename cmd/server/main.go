package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"go.stevenxie.me/api/auth"
	"go.stevenxie.me/api/auth/airtable"

	"go.stevenxie.me/api/productivity/prodsvc"

	"go.stevenxie.me/api/productivity"
	"go.stevenxie.me/api/productivity/rescuetime"

	"go.stevenxie.me/api/git/gitgh"
	"go.stevenxie.me/api/git/gitsvc"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli"

	"go.stevenxie.me/gopkg/cmdutil"
	"go.stevenxie.me/gopkg/configutil"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/guillotine"

	"go.stevenxie.me/api/git"
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/location/geocode"
	"go.stevenxie.me/api/location/geocode/here"
	"go.stevenxie.me/api/location/gmaps"
	"go.stevenxie.me/api/location/locsvc"

	"go.stevenxie.me/api/about"
	"go.stevenxie.me/api/about/aboutgh"
	"go.stevenxie.me/api/about/aboutsvc"

	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/music/musicsvc"
	"go.stevenxie.me/api/music/spotify"

	"go.stevenxie.me/api/scheduling"
	"go.stevenxie.me/api/scheduling/gcal"
	"go.stevenxie.me/api/scheduling/schedsvc"

	"go.stevenxie.me/api/pkg/github"
	"go.stevenxie.me/api/pkg/google"
	"go.stevenxie.me/api/pkg/poll"
	"go.stevenxie.me/api/pkg/svcutil"
	"go.stevenxie.me/api/server/httpsrv"

	"go.stevenxie.me/api/cmd/server/config"
	cmdint "go.stevenxie.me/api/cmd/server/internal"
	"go.stevenxie.me/api/internal"
)

func main() {
	// Load envvars from dotenv.
	if err := configutil.LoadEnv(); err != nil {
		cmdutil.Fatalf("Failed to load dotenv file: %v\n", err)
	}

	app := cli.NewApp()
	app.Name = cmdint.Name
	app.Usage = "An server for my personal API."
	app.UsageText = fmt.Sprintf("%s [global options]", cmdint.Name)
	app.Version = internal.Version
	app.Action = run

	// Hide help command.
	app.Commands = []cli.Command{{Name: "help", Hidden: true}}

	// Configure flags.
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "port",
			Usage:       "port that the HTTP server listens on",
			Value:       3000,
			Destination: &flags.HTTPPort,
		},
		cli.BoolFlag{
			Name:  "help,h",
			Usage: "show help",
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		os.Exit(1)
	}
}

var flags struct {
	HTTPPort int
}

func run(*cli.Context) (err error) {
	// Init logger, and Raven client.
	var (
		raven = buildRaven()
		log   = buildLogger(raven)
	)

	// Load and validate config.
	cfg, err := config.Load()
	if err != nil {
		return errors.Wrap(err, "loading config")
	}
	if err = cfg.Validate(); err != nil {
		return errors.Wrap(err, "invalid config")
	}

	// Init guillotine.
	guillo := guillotine.New(guillotine.WithLogger(
		logutil.WithComponent(log, "guillotine.Guillotine"),
	))
	guillo.TriggerOnTerminate()
	defer func() {
		if ok, _ := guillo.Execute(); !ok && (err != nil) {
			err = errors.New("guillotine finished running with errors")
		}
	}()

	// Connect to data sources.
	log.Info("Connecting to data sources...")

	timelineClient, err := gmaps.NewTimelineClient()
	if err != nil {
		return errors.Wrap(err, "creating Google Maps timeline client")
	}

	hereClient, err := here.NewClient(cfg.Location.Here.AppID)
	if err != nil {
		return errors.Wrap(err, "creating Here client")
	}

	githubClient, err := github.New()
	if err != nil {
		return errors.Wrap(err, "creating GitHub client")
	}

	spotifyClient, err := spotify.New()
	if err != nil {
		return errors.Wrap(err, "creating Spotify client")
	}

	googleClients, err := google.NewClientSet()
	if err != nil {
		return errors.Wrap(err, "creating Google client set")
	}

	rtimeClient, err := rescuetime.NewClient()
	if err != nil {
		return errors.Wrap(err, "creating RescueTime client")
	}

	airtableClient, err := airtable.NewClient()
	if err != nil {
		return errors.Wrap(err, "creating Airtable client")
	}

	// Init services.
	log.Info("Initializing services...")

	var locationService location.Service
	{
		var (
			source         = gmaps.NewSegmentSource(timelineClient)
			geocoder       = here.NewGeocoder(hereClient)
			historyService = locsvc.NewHistoryService(
				source, geocoder,
				svcutil.WithLogger(logutil.WithComponent(log, "location.HistoryService")),
			)
		)

		if cfg := cfg.Location.Precacher; cfg.Enabled {
			historyPrecacher := locsvc.NewHistoryServicePrecacher(
				historyService,
				cfg.Interval,
				poll.WithPrecacherLogger(
					logutil.WithComponent(log, "locsvc.HistoryServicePrecacher"),
				),
			)
			guillo.AddFunc(
				historyPrecacher.Stop,
				guillotine.WithPrefix("stopping location history service precacher"),
			)
			historyService = historyPrecacher
		}

		geocodeLevel, err := geocode.ParseLevel(
			cfg.Location.CurrentRegion.GeocodeLevel,
		)
		if err != nil {
			return errors.Wrap(err, "parsing geocode level")
		}
		locationService = locsvc.NewService(
			historyService, geocoder,
			locsvc.WithLogger(logutil.WithComponent(log, "location.Service")),
			locsvc.WithRegionGeocodeLevel(geocodeLevel),
		)
	}

	var aboutService about.Service
	{
		var (
			gist   = cfg.About.Gist
			source = aboutgh.NewStaticSource(
				githubClient.GitHub().Gists,
				gist.ID, gist.File,
			)
		)
		aboutService = aboutsvc.NewService(
			source, locationService,
			svcutil.WithLogger(logutil.WithComponent(log, "about.Service")),
		)
	}

	var musicService music.Service
	{
		var (
			source        = spotify.NewSource(spotifyClient)
			sourceService = musicsvc.NewSourceService(
				source,
				svcutil.WithLogger(logutil.WithComponent(log, "music.SourceService")),
			)
		)
		var (
			currentSource  = spotify.NewCurrentSource(spotifyClient)
			currentService = musicsvc.NewCurrentService(
				currentSource,
				svcutil.WithLogger(logutil.WithComponent(log, "music.CurrentService")),
			)
		)
		var (
			controller     = spotify.NewController(spotifyClient)
			controlService = musicsvc.NewControlService(
				controller,
				svcutil.WithLogger(logutil.WithComponent(log, "music.ControlService")),
			)
		)
		musicService = musicsvc.NewService(
			sourceService,
			currentService,
			controlService,
		)
	}

	var musicStreamer music.Streamer
	{
		currentStreamer := musicsvc.NewCurrentStreamer(
			musicService,
			musicsvc.WithCurrentStreamerPollInterval(cfg.Music.Streamer.PollInterval),
			musicsvc.WithCurrentStreamerLogger(
				logutil.WithComponent(log, "music.CurrentStreamer"),
			),
		)
		guillo.AddFunc(
			currentStreamer.Stop,
			guillotine.WithPrefix("stopping music streamer"),
		)
		musicStreamer = currentStreamer
	}

	var schedulingService scheduling.Service
	{
		calsvc, err := googleClients.CalendarService(context.Background())
		if err != nil {
			return errors.Wrap(err, "creating Google calendar service")
		}
		source := gcal.NewBusySource(
			calsvc,
			cfg.Scheduling.GCal.CalendarIDs,
		)
		schedulingService = schedsvc.NewService(
			source,
			svcutil.WithLogger(logutil.WithComponent(log, "scheduling.Service")),
		)
	}

	var gitService git.Service
	{
		source := gitgh.NewSource(githubClient)
		gitService = gitsvc.NewService(
			source,
			svcutil.WithLogger(logutil.WithComponent(log, "git.Service")),
		)

		if cfg := cfg.Git.Precacher; cfg.Enabled {
			precacher := gitsvc.NewServicePrecacher(
				gitService,
				cfg.Interval,
				func(spCfg *gitsvc.ServicePrecacherConfig) {
					spCfg.Logger = logutil.WithComponent(log, "gitsvc.ServicePrecacher")
					if l := cfg.Limit; l != nil {
						spCfg.Limit = l
					}
				},
			)
			guillo.AddFunc(
				precacher.Stop,
				guillotine.WithPrefix("stopping Git service precacher"),
			)
			gitService = precacher
		}
	}

	var productivityService productivity.Service
	{
		source := rescuetime.NewRecordSource(rtimeClient)
		productivityService = prodsvc.NewService(
			source,
			locationService,
			svcutil.WithLogger(logutil.WithComponent(log, "productivity.Service")),
		)
	}

	var authService auth.Service
	{
		atCfg := cfg.Auth.Airtable
		authService = airtable.NewService(
			airtableClient,
			atCfg.Selector,
			func(svcCfg *airtable.ServiceConfig) {
				if access := atCfg.AccessRecords; access.Enabled {
					svcCfg.AccessSelector = &access.Selector
					svcCfg.Logger = logutil.WithComponent(log, "auth.Service")
				}
			},
		)
	}

	// Start HTTP server.
	log.Info("Initializing HTTP server...")
	srv := httpsrv.NewServer(
		httpsrv.Services{
			Git:          gitService,
			About:        aboutService,
			Music:        musicService,
			Auth:         authService,
			Location:     locationService,
			Scheduling:   schedulingService,
			Productivity: productivityService,
		},
		httpsrv.Streamers{
			Music: musicStreamer,
		},
		httpsrv.WithLogger(logutil.WithComponent(log, "httpsrv.Server")),
	)
	guillo.AddFinalizer(func() error {
		var (
			log = log
			ctx = context.Background()
		)
		if timeout := cfg.Server.ShutdownTimeout; timeout != nil {
			log = log.WithField("timeout", *timeout)
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, *timeout)
			defer cancel()
		}
		log.Info("Shutting down HTTP server...")
		err := srv.Shutdown(ctx)
		return errors.Wrap(err, "shutting down server")
	})

	// Listen for new connections.
	err = srv.ListenAndServe(fmt.Sprintf(":%d", flags.HTTPPort))
	if !errors.Is(err, http.ErrServerClosed) {
		guillo.Trigger()
		return errors.Wrap(err, "starting HTTP server")
	}
	return nil
}
