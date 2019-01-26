package server

import (
	"github.com/stevenxie/api/pkg/data/airtable"
	"github.com/stevenxie/api/pkg/mood"
	"github.com/stevenxie/api/pkg/mood/adapter"
	ess "github.com/unixpickle/essentials"
)

// startCron starts the Server's cron manager.
func (s *Server) startCron() {
	for _, entry := range s.Cron.Entries() {
		entry.Job.Run()
	}
	s.Cron.Run()
}

func (s *Server) registerCronJobs() error {
	// Configure mood fetcher.
	client, err := airtable.NewUsing(s.viper)
	if err != nil {
		return ess.AddCtx("creating Airtable client", err)
	}

	// Construct adapters.
	var (
		adapter = adapter.AirtableAdapter{Client: client}
		fetcher = mood.NewFetcher(adapter, s.Repos.MoodRepo,
			s.l.Named("moodfetcher"))
	)
	err = fetcher.RegisterTo(s.Cron)
	return ess.AddCtx("registering mood fetcher", err)
}
