package mood

import (
	"errors"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const (
	fetchLimit = 10
	cronSpec   = "@every 15m"
)

// A CronJob is a runnable object.
type CronJob interface{ Run() }

// A CronMan manages CronJobs.
type CronMan interface {
	AddJob(spec string, job CronJob)
}

// A Fetcher is a CronJob that can fetch moods from a source.
type Fetcher struct {
	Source
	*gorm.DB

	l *zap.SugaredLogger
}

// NewFetcher returns a new Fetcher.
func NewFetcher(s Source, db *gorm.DB, l *zap.SugaredLogger) *Fetcher {
	if l == nil {
		l = zap.NewNop().Sugar()
	}
	if db == nil {
		panic(errors.New("cannot create fetcher with nil db"))
	}

	return &Fetcher{
		Source: s,
		DB:     db,
		l:      l,
	}
}

// Run fetches new moods from f.Source, and saves them to f.Repo.
func (f *Fetcher) Run() {
	f.l.Info("Starting fetch procedure...")

	var latest *Mood
	f.DB.Last(&latest)
	if f.DB.Error != nil {
		f.l.Errorf("Failed to fatch last mood from DB: %v", f.DB.Error)
		return
	}

	moods, err := f.Source.FetchMoods(fetchLimit)
	if err != nil {
		f.l.Errorf("Failed to fetch new moods: %v", err)
		return
	}

	// Filter out already-saved moods.
	var keep []*Mood
	if latest == nil {
		keep = moods // keep all moods
	} else {
		for i := range moods {
			if moods[i].ExtID > latest.ExtID {
				keep = append(keep, moods[i])
			}
		}
	}

	// Save new moods.
	f.DB.Save(&keep)
	if f.DB.Error != nil {
		f.l.Errorf("Failed to save new moods to repo: %v", f.DB.Error)
		return
	}
	f.l.Infof("Successfully saved %d new moods to repo.", len(keep))
}
