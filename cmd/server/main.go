package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/getsentry/sentry-go"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"go.stevenxie.me/gopkg/cmdutil"
	"go.stevenxie.me/gopkg/configutil"
	"go.stevenxie.me/guillotine"

	"go.stevenxie.me/api/pkg/basic"
	"go.stevenxie.me/api/pkg/github"
	"go.stevenxie.me/api/pkg/google"
	"go.stevenxie.me/api/pkg/here"
	"go.stevenxie.me/api/pkg/jaeger"

	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/location/geocode"
	"go.stevenxie.me/api/location/geocode/heregeo"
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

	"go.stevenxie.me/api/productivity"
	"go.stevenxie.me/api/productivity/prodsvc"
	"go.stevenxie.me/api/productivity/rescuetime"

	"go.stevenxie.me/api/git"
	"go.stevenxie.me/api/git/gitgh"
	"go.stevenxie.me/api/git/gitsvc"

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/assist/transit/grt"
	"go.stevenxie.me/api/assist/transit/heretrans"
	"go.stevenxie.me/api/assist/transit/transvc"

	"go.stevenxie.me/api/auth"
	"go.stevenxie.me/api/auth/airtable"

	"go.stevenxie.me/api/cmd/server/config"
	cmdint "go.stevenxie.me/api/cmd/server/internal"
	"go.stevenxie.me/api/internal"
	"go.stevenxie.me/api/server/httpsrv"
)

func main() {
	// Load envvars from dotenv.
	if err := configutil.LoadEnv(); err != nil {
		cmdutil.Fatalf("Failed to load dotenv file: %v\n", err)
	}
	app := cli.NewApp()
	app.Name = cmdint.Name
	app.Usage = "A server for my personal API."
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
	// Init logger, and Sentry client.
	//
	// TODO: Use Logrus hook that uses the sentry-go client.
	var (
		log    *logrus.Entry
		sentry *sentry.Client
	)
	{
		opt := cmdutil.WithRelease(internal.Version)
		raven := cmdutil.NewRaven(opt)
		sentry = cmdutil.NewSentry(opt)
		log = cmdutil.NewLogger(cmdutil.WithSentryHook(raven))
	}

	// Load and validate config.
	cfg, err := config.Load()
	if err != nil {
		return errors.Wrap(err, "loading config")
	}
	if err = cfg.Validate(); err != nil {
		return errors.Wrap(err, "invalid config")
	}

	// Init guillotine.
	guillo := guillotine.New(guillotine.WithLogger(log))
	guillo.TriggerOnTerminate()
	defer func() {
		if ok, _ := guillo.Execute(); !ok && (err != nil) {
			err = errors.New("guillotine finished running with errors")
		}
	}()

	// Init tracer.
	var tracer opentracing.Tracer
	if t := cfg.Tracer; t.Enabled {
		var closer io.Closer
		tracer, closer, err = jaeger.NewTracer(
			cmdint.Namespace,
			func(cfg *jaeger.Config) {
				jaeger.MergeConfigs(cfg, &t.Jaeger)
			},
		)
		if err != nil {
			return errors.Wrap(err, "creating Jaeger tracer")
		}
		guillo.AddCloser(closer, guillotine.WithPrefix("closing Jaeger tracer"))
	} else {
		tracer = new(opentracing.NoopTracer)
	}

	// Connect to data sources.
	log.Info("Connecting to data sources...")

	timelineClient, err := gmaps.NewTimelineClient()
	if err != nil {
		return errors.Wrap(err, "create Google Maps timeline client")
	}

	hereClient, err := here.NewClient(cfg.Location.Here.AppID)
	if err != nil {
		return errors.Wrap(err, "create Here client")
	}

	githubClient, err := github.New()
	if err != nil {
		return errors.Wrap(err, "create GitHub client")
	}

	spotifyClient, err := spotify.New()
	if err != nil {
		return errors.Wrap(err, "create Spotify client")
	}

	googleClients, err := google.NewClientSet()
	if err != nil {
		return errors.Wrap(err, "create Google client set")
	}

	rtimeClient, err := rescuetime.NewClient()
	if err != nil {
		return errors.Wrap(err, "create RescueTime client")
	}

	airtableClient, err := airtable.NewClient()
	if err != nil {
		return errors.Wrap(err, "create Airtable client")
	}

	// Init services.
	log.Info("Initializing services...")

	var locationService location.Service
	{
		var (
			geoc    = heregeo.NewGeocoder(hereClient, basic.WithTracer(tracer))
			hist    = gmaps.NewHistorian(timelineClient, basic.WithTracer(tracer))
			histsvc = locsvc.NewHistoryService(
				hist, geoc,
				basic.WithLogger(log),
				basic.WithTracer(tracer),
			)
		)

		if cfg := cfg.Location.Precacher; cfg.Enabled {
			historyPrecacher := locsvc.NewHistoryServicePrecacher(
				histsvc,
				cfg.Interval,
				basic.WithLogger(log),
			)
			guillo.AddFunc(
				historyPrecacher.Stop,
				guillotine.WithPrefix("stopping location.HistoryServicePrecacher"),
			)
			histsvc = historyPrecacher
		}

		geocodeLevel, err := geocode.ParseLevel(
			cfg.Location.CurrentRegion.GeocodeLevel,
		)
		if err != nil {
			return errors.Wrap(err, "parsing geocode level")
		}
		locationService = locsvc.NewService(
			histsvc, geoc,
			locsvc.WithLogger(log),
			locsvc.WithTracer(tracer),
			locsvc.WithRegionGeocodeLevel(geocodeLevel),
		)
	}

	var aboutService about.Service
	{
		var (
			gist = cfg.About.Gist
			src  = aboutgh.NewStaticSource(
				githubClient.GitHub().Gists,
				gist.ID, gist.File,
			)
		)
		aboutService = aboutsvc.NewService(
			src, locationService,
			basic.WithLogger(log),
			basic.WithTracer(tracer),
		)
	}

	var musicService music.Service
	{
		var (
			src            = spotify.NewSource(spotifyClient)
			srcsvc         = musicsvc.NewSourceService(src, basic.WithLogger(log))
			currentService = spotify.NewCurrentService(
				spotifyClient,
				basic.WithLogger(log),
				basic.WithTracer(tracer),
			)
		)
		var (
			ctrl    = spotify.NewController(spotifyClient)
			ctrlsvc = musicsvc.NewControlService(
				ctrl,
				basic.WithLogger(log),
				basic.WithTracer(tracer),
			)
		)
		musicService = musicsvc.NewService(
			srcsvc,
			currentService,
			ctrlsvc,
		)
	}

	var musicStreamer music.Streamer
	if cfg := cfg.Music.Streamer; cfg.Enabled {
		currentStreamer := musicsvc.NewCurrentStreamer(
			musicService,
			musicsvc.WithCurrentStreamerPollInterval(cfg.PollInterval),
			musicsvc.WithCurrentStreamerLogger(log),
		)
		guillo.AddFunc(
			currentStreamer.Stop,
			guillotine.WithPrefix("stopping music streamer"),
		)
		musicStreamer = currentStreamer
	} else {
		musicStreamer = musicsvc.NewNoopCurrentStreamer(basic.WithLogger(log))
	}

	var schedulingService scheduling.Service
	{
		calsvc, err := googleClients.CalendarService(context.Background())
		if err != nil {
			return errors.Wrap(err, "create Google calendar service")
		}
		src := gcal.NewCalendar(calsvc, cfg.Scheduling.GCal.CalendarIDs)
		schedulingService = schedsvc.NewService(
			src,
			locationService,
			basic.WithLogger(log),
			basic.WithTracer(tracer),
		)
	}

	var gitService git.Service
	{
		src := gitgh.NewSource(githubClient)
		gitService = gitsvc.NewService(
			src,
			basic.WithLogger(log),
			basic.WithTracer(tracer),
		)

		if pc := cfg.Git.Precacher; pc.Enabled {
			precacher := gitsvc.NewServicePrecacher(
				gitService,
				pc.Interval,
				func(cfg *gitsvc.ServicePrecacherConfig) {
					cfg.Logger = log
					if l := pc.Limit; l != nil {
						cfg.Limit = l
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
		src := rescuetime.NewRecordSource(rtimeClient)
		productivityService = prodsvc.NewService(
			src,
			locationService,
			basic.WithLogger(log),
			basic.WithTracer(tracer),
		)
	}

	var authService auth.Service
	{
		at := cfg.Auth.Airtable
		authService = airtable.NewService(
			airtableClient,
			at.Codes.Selector,
			airtable.WithLogger(log),
			airtable.WithTracer(tracer),
			func(cfg *airtable.ServiceConfig) {
				if access := at.AccessRecords; access.Enabled {
					cfg.AccessSelector = &access.Selector
				}
			},
		)
	}

	var transitService transit.Service
	{
		var (
			loc    = heretrans.NewLocator(hereClient)
			locsvc = transvc.NewLocatorService(
				loc,
				basic.WithLogger(log),
				basic.WithTracer(tracer),
			)
		)
		grt, err := grt.NewRealtimeSource(
			grt.WithLogger(log),
			grt.WithTracer(tracer),
		)
		if err != nil {
			return errors.Wrap(err, "create grt.RealTimeSource")
		}
		transitService = transvc.NewService(
			locsvc,
			transvc.WithLogger(log),
			transvc.WithTracer(tracer),
			transvc.WithRealtimeSource(grt, transit.OpCodeGRT),
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
			Transit:      transitService,
			Scheduling:   schedulingService,
			Productivity: productivityService,
		},
		httpsrv.Streamers{
			Music: musicStreamer,
		},
		httpsrv.WithLogger(log),
		httpsrv.WithSentry(sentry),
	)
	guillo.AddFinalizer(func() error {
		var (
			log = log
			ctx = context.Background()
		)
		if timeout := cfg.ShutdownTimeout; timeout != nil {
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
