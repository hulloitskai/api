package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/pflag"
	ess "github.com/unixpickle/essentials"

	"github.com/stevenxie/api/internal/config"
	"github.com/stevenxie/api/internal/info"
	"github.com/stevenxie/api/jobserver"
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
	l, err := config.BuildLogger()
	if err != nil {
		ess.Die("Error while building zap.SugaredLogger:", err)
	}

	// Create viper instance.
	v := config.BuildViper()
	if opts.ConfigPath != "" {
		v.AddConfigPath(opts.ConfigPath)
	}
	if err = v.ReadInConfig(); err != nil {
		l.Fatalf("Error while reading Viper config: %v", err)
	}

	// Create data providers.
	l.Info("Creating service provider...")
	provider, err := newProvider(v)
	if err != nil {
		l.Fatalf("Error while creating service provider: %v", err)
	}
	if err = provider.Open(); err != nil {
		l.Fatalf("Error while initializing provider: %v", err)
	}
	defer func() {
		if err = provider.Close(); err != nil {
			l.Errorf("Error while closing Mongo: %v", err)
		}
	}()

	// Create and start server.
	srv := jobserver.NewViper(provider, v.Sub(jobserver.Namespace))
	srv.SetLogger(l.Named("jobserver"))
	l.Info("Starting job server...")
	if err = srv.Start(); err != nil {
		l.Errorf("Error while starting job server: %v", err)
	}

	// Wait for kill / interrupt signal.
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)

	<-ch
	l.Info("Shutting down server gracefully...")
	if err := srv.Stop(); err != nil {
		l.Fatalf("Error while stopping manager: %v", err)
	}
}
