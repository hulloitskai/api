package cli

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/spf13/pflag"
	"github.com/stevenxie/api/internal/info"
	"github.com/stevenxie/api/internal/server"
	ess "github.com/unixpickle/essentials"
)

// opts are a set of program options.
var opts struct {
	ShowVersion bool
	ShowHelp    bool
	Port        int
}

// Define CLI flags, initialize program.
func init() {
	pflag.BoolVarP(&opts.ShowHelp, "help", "h", false, "Show help (usage).")
	pflag.BoolVarP(&opts.ShowVersion, "version", "v", false, "Show version.")
	pflag.IntVarP(&opts.Port, "port", "p", 3000, "Port to listen on.")

	loadEnv()     // load .env variables
	pflag.Parse() // parse CLI arguments
}

// Exec is the entrypoint to command rgv.
func Exec() {
	if opts.ShowHelp {
		pflag.Usage()
		os.Exit(0)
	}
	if opts.ShowVersion {
		fmt.Println(info.Version)
		os.Exit(0)
	}

	// Create program logger.
	logger, err := buildLogger()
	if err != nil {
		ess.Die("Error while building zap.SugaredLogger:", err)
	}

	// Create and run server.
	s, err := server.New(logger)
	if err != nil {
		ess.Die("Error while building server:", err)
	}

	addr := fmt.Sprintf(":%d", opts.Port)
	fmt.Printf("Listening on address '%s'...\n", addr)

	go shutdownUponInterrupt(s)
	if err = s.ListenAndServe(addr); (err != nil) &&
		(err != http.ErrServerClosed) {
		ess.Die("Error while starting server:", err)
	}
}

func shutdownUponInterrupt(s *server.Server) {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)

	<-ch // wait for a signal
	fmt.Printf("Shutting down server gracefully (timeout: %v)...\n",
		s.ShutdownTimeout)
	if err := s.Shutdown(); err != nil {
		ess.Die("Error during server shutdown:", err)
	}
}
