package config

import (
	"github.com/stevenxie/api/pkg/gitutil"
	"github.com/stevenxie/api/server"
)

// AboutGistInfo returns info required for creating a new github.AboutService.
func (cfg *Config) AboutGistInfo() (id, file string) {
	gist := &cfg.About.Gist
	return gist.ID, gist.File
}

// CommitLoaderOpts returns options used to configure a gitutil.CommitLoader.
func (cfg *Config) CommitLoaderOpts() []gitutil.CLOption {
	var (
		opts    []gitutil.CLOption
		commits = &cfg.Commits
	)
	if commits.Limit != nil {
		opts = append(opts, gitutil.WithLimit(*commits.Limit))
	}
	if commits.PollInterval != nil {
		opts = append(opts, gitutil.WithInterval(*commits.PollInterval))
	}
	return opts
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
		nowPlaying = &cfg.NowPlaying
	)
	if nowPlaying.PollInterval != nil {
		opts = append(
			opts,
			server.WithNowPlayingPollInterval(*nowPlaying.PollInterval),
		)
	}
	return opts
}
