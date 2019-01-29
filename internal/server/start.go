package server

import (
	"net/http"

	"github.com/stevenxie/api/internal/data"
	"github.com/stevenxie/api/internal/server/routes"
	ess "github.com/unixpickle/essentials"
)

// ListenAndServe starts the server on the specified address.
func (s *Server) ListenAndServe(addr string) error {
	// Load data drivers and repositories.
	s.l.Info("Loading data drivers...")
	drivers, err := data.LoadDrivers(s.viper)
	if err != nil {
		return ess.AddCtx("server: loading drivers", err)
	}
	s.drivers = drivers
	if s.Repos, err = data.NewRepoSet(drivers); err != nil {
		return ess.AddCtx("server: creating repos", err)
	}

	// Register and start cron jobs.
	s.l.Info("Registering cron jobs...")
	if err = s.registerCronJobs(); err != nil {
		return ess.AddCtx("server: registering cron jobs", err)
	}
	go s.startCron()

	// Make and configure router.
	s.l.Info("Registering routes...")
	router, err := routes.NewRouter(s.Repos, s.l.Named("router"))
	if err != nil {
		return ess.AddCtx("server: creating router", err)
	}

	s.WebServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}
	s.l.Infof("Listening on address '%s'...", addr)
	return s.WebServer.ListenAndServe()
}
