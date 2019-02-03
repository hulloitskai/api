package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/pflag"
	ess "github.com/unixpickle/essentials"

	"github.com/stevenxie/api/data/mongo"
	"github.com/stevenxie/api/internal/config"
	"github.com/stevenxie/api/internal/info"
	"github.com/stevenxie/api/jobserver"
	"github.com/stevenxie/api/work"
	"github.com/stevenxie/api/work/airtable"
)

// opts are a set of program options.
var opts struct {
	ShowVersion bool
	ShowHelp    bool
	ConfigPath  string
}

// Define CLI flags, initialize program.
func init() {
	if err := config.LoadDotEnv(); err != nil {
		ess.Die("Reading '.env' files:", err)
	}

	pflag.BoolVarP(&opts.ShowHelp, "help", "h", false, "Show help (usage).")
	pflag.BoolVarP(&opts.ShowVersion, "version", "v", false, "Show version.")
	pflag.StringVarP(&opts.ConfigPath, "config", "c", "", "Path to config file.")

	pflag.Parse() // parse CLI arguments
}

func main() {
	if opts.ShowHelp {
		pflag.Usage()
		os.Exit(0)
	}
	if opts.ShowVersion {
		fmt.Println(info.Version)
		os.Exit(0)
	}

	// Create program logger.
	logger, err := config.BuildLogger()
	if err != nil {
		ess.Die("Error while building zap.SugaredLogger:", err)
	}

	// Create viper instance.
	v := config.BuildViper()
	if opts.ConfigPath != "" {
		v.AddConfigPath(opts.ConfigPath)
	}
	if err = v.ReadInConfig(); err != nil {
		ess.Die("Error while reading Viper config:", err)
	}

	// Create data providers.
	fmt.Println("Creating data providers...")
	ap, err := airtable.NewUsing(v)
	if err != nil {
		ess.Die("Error while creating Airtable provider:", err)
	}
	// Create mongo provider.
	mp, err := mongo.NewProviderUsing(v)
	if err != nil {
		ess.Die("Error while creating Mongo provider:", err)
	}
	if err := mp.Open(); err != nil {
		ess.Die("Error while opening Mongo provider:", err)
	}

	// Create server.
	s, err := jobserver.NewUsing(v)
	if err != nil {
		ess.Die("Error while creating manager:", err)
	}

	// Install handlers.
	mf := work.NewMoodFetcher(ap, mp.MoodService, logger.Named("moodfetcher"))
	s.RegisterMoodFetcher(mf)

	// Start manager, shutdown upon interrupt.
	fmt.Println("Starting work server...")
	s.Start()
	shutdownUponInterrupt(s, mp)
}

func shutdownUponInterrupt(s *jobserver.Server, mp *mongo.Provider) {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)

	<-ch // wait for a signal
	fmt.Println("Shutting down server gracefully...")
	if err := s.Stop(); err != nil {
		ess.Die("Error while stopping manager:", err)
	}
	if err := mp.Close(); err != nil {
		ess.Die("Error while closing Mongo:", err)
	}
}
