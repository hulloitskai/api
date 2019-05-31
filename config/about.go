package config

import (
	"github.com/stevenxie/api/data/github"
)

// BuildInfoStore builds a preconfigured about.InfoStore.
func (cfg *Config) BuildInfoStore(gr github.GistRepo) *github.InfoStore {
	gist := &cfg.About.Gist
	return github.NewInfoStore(gr, gist.ID, gist.File)
}
