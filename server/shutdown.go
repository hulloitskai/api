package server

// Shutdown gracefully shuts down the Server.
func (s *Server) Shutdown() error {
	ctx, cancel := s.Config.ShutdownContext()
	defer cancel()
	return s.WebServer.Shutdown(ctx)
}
