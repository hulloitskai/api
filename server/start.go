package server

// ListenAndServe starts the server on the specified address.
func (s *Server) ListenAndServe(addr string) error {
	s.WebServer.Addr = addr
	s.l.Infof("Listening on address '%s'...", addr)
	return s.WebServer.ListenAndServe()
}
