package mood

import (
	"github.com/robfig/cron"
	"go.uber.org/zap"
)

const (
	fetchLimit = 10
	cronSpec   = "@every 15m"
)

// A Fetcher is a CronJob that can fetch moods from a source.
type Fetcher struct {
	Source Source
	Repo   Repo

	l *zap.SugaredLogger
}

// NewFetcher returns a new Fetcher.
func NewFetcher(s Source, r Repo, l *zap.SugaredLogger) *Fetcher {
	if l == nil {
		l = zap.NewNop().Sugar()
	}
	return &Fetcher{
		Source: s,
		Repo:   r,
		l:      l,
	}
}

// Run fetches new moods from f.Source, and saves them to f.Repo.
func (f *Fetcher) Run() {
	f.l.Info("Starting fetch procedure...")

	// Select last mood from repo.
	var (
		results, err = f.Repo.SelectMoods(1, "")
		last         *Mood
	)
	if err != nil {
		f.l.Errorf("Error while selecting last mood: %v", err)
		return
	}
	if len(results) > 0 {
		last = results[0]
		f.l.Debugf("Last saved mood: %+v", last)
	}

	// Fetch new moods from source.
	moods, err := f.Source.FetchMoods(fetchLimit)
	if err != nil {
		f.l.Errorf("Failed to fetch new moods: %v", err)
		return
	}

	// Filter out already-saved moods.
	var keep []*Mood
	if last == nil {
		keep = moods // keep all moods
	} else {
		for i := range moods {
			if moods[i].ExtID > last.ExtID {
				keep = append(keep, moods[i])
			}
		}
	}
	if len(keep) == 0 {
		f.l.Infof("No new moods available to save.")
		return
	}

	// Save new moods to repo.
	if err := f.Repo.InsertMoods(keep); err != nil {
		f.l.Errorf("Failed to insert moods: %v", err)
		return
	}
	f.l.Infof("Successfully saved %d new moods to repo.", len(keep))
}

// Cron manages CronJobs.
type Cron interface {
	AddJob(spec string, job cron.Job) error
}

// RegisterTo registers the Fetcher to a Cron.
func (f *Fetcher) RegisterTo(c Cron) error {
	return c.AddJob(cronSpec, f)
}
