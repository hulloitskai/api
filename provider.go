package api

// ServiceProvider provides the underlying services required by a Server.
//
// It contains connections to external services (i.e. databases) that can be
// opened, and closed.
type ServiceProvider interface {
	Open() error
	Close() error

	MoodService
}
