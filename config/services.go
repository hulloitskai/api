package config

// AboutGistInfo returns info required for creating a new github.AboutService.
func (cfg *Config) AboutGistInfo() (id, file string) {
	gist := &cfg.About.Gist
	return gist.ID, gist.File
}

// GCalCalendarIDs returns the calendar IDs required for creating a new
// gcal.Client.
func (cfg *Config) GCalCalendarIDs() []string {
	return cfg.Availability.GCal.CalendarIDs
}
