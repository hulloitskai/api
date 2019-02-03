package work

import (
	"go.uber.org/zap"

	"github.com/stevenxie/api"
	ess "github.com/unixpickle/essentials"
)

// A MoodSource can be polled for new moods.
type MoodSource interface{ GetNewMoods() ([]*api.Mood, error) }

// A MoodFetcher can fetch moods from a MoodSource, and save them into a
// MoodService.
type MoodFetcher struct {
	Src MoodSource
	Svc api.MoodService
	l   *zap.SugaredLogger
}

// NewMoodFetcher builds a MoodFetcher.
func NewMoodFetcher(src MoodSource, svc api.MoodService,
	l *zap.SugaredLogger) *MoodFetcher {
	if l == nil {
		l = zap.NewNop().Sugar()
	}
	return &MoodFetcher{
		Src: src,
		Svc: svc,
		l:   l,
	}
}

// FetchMoods fetches new moods from mf.Src, and places them in mf.Svc.
func (mf *MoodFetcher) FetchMoods() error {
	mf.l.Info("Fetching new moods...")

	// Fetch new moods from source.
	nmoods, err := mf.Src.GetNewMoods()
	if err != nil {
		mf.l.Errorf("Error getting new moods from source: %v", err)
		return ess.AddCtx("job: getting new moods from source", err)
	}
	mf.l.Debugf("Got new moods from source: %+v", nmoods)

	// Get last saved mood.
	lmoods, err := mf.Svc.ListMoods(1, 0)
	if err != nil {
		mf.l.Errorf("Error getting last moods from service: %v", err)
		return ess.AddCtx("job: getting last moods from service", err)
	}

	// Filter out already-saved moods.
	keep := make([]*api.Mood, 0, len(nmoods))
	if len(lmoods) == 0 {
		keep = nmoods
	} else {
		for _, mood := range nmoods {
			if mood.ExtID > lmoods[0].ExtID {
				keep = append(keep, mood)
			}
		}
	}
	if len(keep) == 0 {
		mf.l.Infof("No new moods to save.")
		return nil
	}

	// Save new moods to service.
	if err := mf.Svc.CreateMoods(keep); err != nil {
		mf.l.Errorf("Error while creating moods in service: %v", err)
		return ess.AddCtx("job: creating moods in service", err)
	}
	mf.l.Infof("Saved %d new moods.", len(keep))
	return nil
}
