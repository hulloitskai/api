package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cockroachdb/errors"
	sentry "github.com/getsentry/sentry-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"go.stevenxie.me/gopkg/cmdutil"
	"go.stevenxie.me/gopkg/configutil"
	"go.stevenxie.me/guillotine"

	"go.stevenxie.me/api/v2/pkg/basic"
	"go.stevenxie.me/api/v2/pkg/github"
	"go.stevenxie.me/api/v2/pkg/google"
	"go.stevenxie.me/api/v2/pkg/here"
	"go.stevenxie.me/api/v2/pkg/jaeger"

	"go.stevenxie.me/api/v2/location"
	"go.stevenxie.me/api/v2/location/geocode"
	"go.stevenxie.me/api/v2/location/geocode/heregeo"
	"go.stevenxie.me/api/v2/location/gmaps"
	"go.stevenxie.me/api/v2/location/locsvc"

	"go.stevenxie.me/api/v2/about"
	"go.stevenxie.me/api/v2/about/aboutgh"
	"go.stevenxie.me/api/v2/about/aboutsvc"

	"go.stevenxie.me/api/v2/music"
	"go.stevenxie.me/api/v2/music/musicsvc"
	"go.stevenxie.me/api/v2/music/spotify"

	"go.stevenxie.me/api/v2/scheduling"
	"go.stevenxie.me/api/v2/scheduling/gcal"
	"go.stevenxie.me/api/v2/scheduling/schedsvc"

	"go.stevenxie.me/api/v2/productivity"
	"go.stevenxie.me/api/v2/productivity/prodsvc"
	"go.stevenxie.me/api/v2/productivity/rescuetime"

	"go.stevenxie.me/api/v2/git"
	"go.stevenxie.me/api/v2/git/gitgh"
	"go.stevenxie.me/api/v2/git/gitsvc"

	"go.stevenxie.me/api/v2/assist/transit"
	"go.stevenxie.me/api/v2/assist/transit/grt"
	"go.stevenxie.me/api/v2/assist/transit/heretrans"
	"go.stevenxie.me/api/v2/assist/transit/transvc"

	"go.stevenxie.me/api/v2/auth"
	"go.stevenxie.me/api/v2/auth/airtable"

	"go.stevenxie.me/api/v2/server/debugsrv"
	"go.stevenxie.me/api/v2/server/gqlsrv"

	"go.stevenxie.me/api/v2/cmd/server/config"
	cmdinternal "go.stevenxie.me/api/v2/cmd/server/internal"
	"go.stevenxie.me/api/v2/internal"
)

func main() {
	// Load envvars from dotenv.
	if err := configutil.LoadEnv(); err != nil {
		cmdutil.Fatalf("Failed to load dotenv file: %v\n", err)
	}
	app := cli.NewApp()
	app.Name = cmdinternal.Name
	app.Usage = "A server for my personal API."
	app.UsageText = fmt.Sprintf("%s [global options]", cmdinternal.Name)
	app.Version = internal.Version
	app.Action = run

	// Hide help command.
	app.Commands = []cli.Command{{Name: "help", Hidden: true}}

	// Configure flags.
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "port,p",
			Usage:       "port that the server listens on",
			Value:       3000,
			Destination: &flags.Port,
		},

		cli.BoolFlag{
			Name:        "debug",
			Usage:       "enable debug server",
			Destination: &flags.Debug,
		},
		cli.IntFlag{
			Name:        "debug-port",
			Usage:       "port that the debug server listens on",
			Value:       6060,
			Destination: &flags.DebugPort,
		},

		cli.DurationFlag{
			Name:        "shutdown-timeout",
			Usage:       "timeout for server shutdown",
			Value:       -1 * time.Second,
			Destination: &flags.ShutdownTimeout,
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
	ShutdownTimeout time.Duration
	Port            int

	Debug     bool
	DebugPort int
}

func run(*cli.Context) (err error) {
	// Init logger, and Sentry client.
	//
	// TODO: Use Logrus hook that uses the sty-go client.
	var (
		log *logrus.Entry
		sty *sentry.Client
	)
	{
		opt := cmdutil.WithRelease(internal.Version)
		raven := cmdutil.NewRaven(opt)
		sty = cmdutil.NewSentry(opt)
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
	if cfg := cfg.Tracer; cfg.Enabled {
		var (
			closer io.Closer
			opts   []jaeger.Option
		)
		{
			j := cfg.Jaeger
			if s := j.Sampler; s != nil {
				opts = append(opts, jaeger.WithSamplerConfig(s))
			}
			if r := j.Reporter; r != nil {
				opts = append(opts, jaeger.WithReporterConfig(r))
			}
			tracer, closer, err = jaeger.NewTracer(cmdinternal.Namespace, opts...)
		}
		if err != nil {
			return errors.Wrap(err, "creating Jaeger tracer")
		}
		guillo.AddCloser(closer, guillotine.WithPrefix("closing Jaeger tracer"))
	} else {
		tracer = new(opentracing.NoopTracer)
	}

	basicOpts := []basic.Option{
		basic.WithLogger(log),
		basic.WithTracer(tracer),
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
			histsvc = locsvc.NewHistoryService(hist, geoc, basicOpts...)
		)

		if cfg := cfg.Location.Precacher; cfg.Enabled {
			historyPrecacher := locsvc.NewHistoryServicePrecacher(
				histsvc,
				cfg.Interval,
				basicOpts...,
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
		aboutService = aboutsvc.NewService(src, locationService, basicOpts...)
	}

	var musicService music.Service
	{
		var (
			src            = spotify.NewSource(spotifyClient)
			srcsvc         = musicsvc.NewSourceService(src, basic.WithLogger(log))
			currentService = spotify.NewCurrentService(spotifyClient, basicOpts...)
		)
		var (
			ctrl    = spotify.NewController(spotifyClient, basicOpts...)
			ctrlsvc = musicsvc.NewControlService(ctrl, basicOpts...)
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
			musicsvc.StreamerWithLogger(log),
			musicsvc.StreamerWithPollInterval(cfg.PollInterval),
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
		schedulingService = schedsvc.NewService(src, locationService, basicOpts...)
	}

	var gitService git.Service
	{
		src := gitgh.NewSource(githubClient)
		gitService = gitsvc.NewService(src, basicOpts...)

		if cfg := cfg.Git.Precacher; cfg.Enabled {
			precacher := gitsvc.NewServicePrecacher(
				gitService,
				cfg.Interval,
				func(opt *gitsvc.ServicePrecacherOptions) {
					opt.Logger = log
					if l := cfg.Limit; l != nil {
						opt.Limit = l
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
		productivityService = prodsvc.NewService(src, locationService, basicOpts...)
	}

	var authService auth.Service
	{
		cfg := cfg.Auth.Airtable
		authService = airtable.NewService(
			airtableClient,
			cfg.Codes.Selector,
			airtable.WithLogger(log),
			airtable.WithTracer(tracer),
			func(opt *airtable.ServiceOptions) {
				if access := cfg.AccessRecords; access.Enabled {
					opt.AccessSelector = &access.Selector
				}
			},
		)
	}

	var transitService transit.Service
	{
		var (
			loc    = heretrans.NewLocator(hereClient)
			locsvc = transvc.NewLocatorService(loc, basicOpts...)
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

	// Coordinate processes with errgroup.
	var group errgroup.Group

	var host string
	if configutil.GetGoEnv() == configutil.GoEnvDevelopment {
		host = "localhost"
	}

	// Start GraphQL server.
	log.Info("Initializing GraphQL server...")
	gqlServer := gqlsrv.NewServer(
		gqlsrv.Services{
			Git:          gitService,
			About:        aboutService,
			Music:        musicService,
			Auth:         authService,
			Location:     locationService,
			Transit:      transitService,
			Scheduling:   schedulingService,
			Productivity: productivityService,
		},
		gqlsrv.Streamers{
			Music: musicStreamer,
		},
		gqlsrv.WithLogger(log),
		gqlsrv.WithSentry(sentry.NewHub(sty, sentry.NewScope())),
	)
	guillo.AddFinalizer(shutdownFinalizer(gqlServer, "GraphQL server", log))
	group.Go(startServerFunc(gqlServer, host, flags.Port, guillo))

	// Start debug server.
	if flags.Debug {
		debugServer := debugsrv.NewServer(basic.WithLogger(log))
		guillo.AddFinalizer(shutdownFinalizer(debugServer, "debug server", log))
		group.Go(startServerFunc(debugServer, host, flags.DebugPort, guillo))
	}

	// Wait for process group to finish.
	return group.Wait()
}

type server interface {
	ListenAndServe(addr string) error
	Shutdown(ctx context.Context) error
}

func shutdownFinalizer(
	srv server, name string,
	log *logrus.Entry,
) guillotine.Finalizer {
	return func() error {
		ctx := context.Background()
		if timeout := flags.ShutdownTimeout; timeout > 0 {
			log = log.WithField("timeout", timeout)
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		log.Infof("Shutting down %s...", name)
		err := srv.Shutdown(ctx)
		return errors.Wrapf(err, "shutdown %s", name)
	}
}

func startServerFunc(
	srv server, host string, port int,
	guillo *guillotine.Guillotine,
) func() error {
	return func() error {
		var (
			addr = fmt.Sprintf("%s:%d", host, port)
			err  = srv.ListenAndServe(addr)
		)
		if !errors.Is(err, http.ErrServerClosed) {
			guillo.Trigger()
			return errors.Wrap(err, "start auth server")
		}
		return nil
	}
}
