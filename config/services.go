package config

import (
	"github.com/stevenxie/api/server"
)

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

// ServerOpts returns options used to configure a server.Server.
func (cfg *Config) ServerOpts() []server.Option {
	var (
		opts       []server.Option
		commits    = &cfg.Commits
		nowPlaying = &cfg.NowPlaying
	)
	if nowPlaying.PollInterval != nil {
		opts = append(
			opts,
			server.WithNowPlayingPollInterval(*nowPlaying.PollInterval),
		)
	}
	if commits.PollInterval != nil {
		opts = append(
			opts,
			server.WithGitCommitsPollInterval(*commits.PollInterval),
		)
	}
	if commits.Limit != nil {
		opts = append(
			opts,
			server.WithGitCommitsLimit(*commits.Limit),
		)
	}
	return opts
}
