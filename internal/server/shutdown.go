package server

import (
	ess "github.com/unixpickle/essentials"
)

// Shutdown gracefully shuts down the Server, closing all existing connections
// and data drivers.
func (s *Server) Shutdown() error {
	// Define shutdown context.
	ctx, cancel := s.Config.ShutdownContext()
	defer cancel()

	// Define shutdown tasks.
	type Task func() error
	tasks := []Task{
		func() error {
			err := s.WebServer.Shutdown(ctx)
			return ess.AddCtx("server: shutting down internal webserver", err)
		},
		func() error {
			s.Cron.Stop()
			return nil
		},
		func() error {
			err := s.drivers.Close()
			return ess.AddCtx("server: closing drivers", err)
		},
	}

	// Perform shutdown.
	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := task(); err != nil {
				return err
			}
		}
	}
	return nil
}
