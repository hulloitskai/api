package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/spf13/pflag"
	ess "github.com/unixpickle/essentials"

	"github.com/stevenxie/api/internal/config"
	"github.com/stevenxie/api/internal/info"
	"github.com/stevenxie/api/server"
)

// opts are a set of program options.
var opts struct {
	ShowVersion bool
	ShowHelp    bool
	Port        int
	ConfigPath  string
}

// Define CLI flags, initialize program.
func init() {
	if err := config.LoadDotEnv(); err != nil {
		ess.Die("Reading '.env' files:", err)
	}

	pflag.BoolVarP(&opts.ShowHelp, "help", "h", false, "Show help (usage).")
	pflag.BoolVarP(&opts.ShowVersion, "version", "v", false, "Show version.")
	pflag.IntVarP(&opts.Port, "port", "p", 3000, "Port to listen on.")
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

	// Create service provider.
	l.Info("Initializing service provider...")
	provider, err := newProvider(v)
	if err != nil {
		l.Fatalf("Error while creating service provider: %v", err)
	}
	if err = provider.Open(); err != nil {
		l.Fatalf("Error while initializing provider: %v", err)
	}
	defer func() {
		l.Info("Closing service provider...")
		if err := provider.Close(); err != nil {
			l.Fatalf("Error while closing service provider: %v", err)
		}
	}()

	// Create and run server.
	srv := server.NewFromViper(provider, v)
	srv.SetLogger(l.Named("server"))
	go func() {
		err := srv.ListenAndServe(fmt.Sprintf(":%d", opts.Port))
		if (err != nil) && (err != http.ErrServerClosed) {
			l.Fatalf("Error while starting server: %v", err)
		}
	}()

	//  Wait for kill / interrupt signal.
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)

	<-ch
	l.Info("Shutting down server gracefully...")
	if err = srv.Shutdown(); err != nil {
		l.Errorf("Error during server shutdown: %v", err)
	}
}
