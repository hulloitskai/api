package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/spf13/pflag"
	ess "github.com/unixpickle/essentials"

	"github.com/stevenxie/api"
	"github.com/stevenxie/api/data"
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

	// Create data provider.
	provider, err := data.NewProviderUsing(v)
	if err != nil {
		ess.Die("Error while creating service provider:", err)
	}

	fmt.Println("Initializing service provider...")
	if err = provider.Open(); err != nil {
		ess.Die("Error while initializing provider:", err)
	}

	// Create and run server.
	cfg, err := server.ConfigFromViper(v)
	if err != nil {
		ess.Die("Error configuring server using Viper:", err)
	}

	s, err := server.New(provider, logger, cfg)
	if err != nil {
		ess.Die("Error while building server:", err)
	}

	fmt.Println("Starting server...")
	addr := fmt.Sprintf(":%d", opts.Port)
	go shutdownUponInterrupt(s, provider)
	if err = s.ListenAndServe(addr); (err != nil) &&
		(err != http.ErrServerClosed) {
		ess.Die("Error while starting server:", err)
	}
}

func shutdownUponInterrupt(s *server.Server, sp api.ServiceProvider) {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)

	<-ch // wait for a signal
	fmt.Printf("Shutting down server gracefully (timeout: %v)...\n",
		s.Config.ShutdownTimeout)
	if err := s.Shutdown(); err != nil {
		ess.Die("Error during server shutdown:", err)
	}

	fmt.Println("Closing data providers...")
	if err := sp.Close(); err != nil {
		ess.Die("Error while closing service provider:", err)
	}
}
