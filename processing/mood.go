package processing

import (
	"github.com/stevenxie/api/internal/util"
	"go.uber.org/zap"
	errors "golang.org/x/xerrors"

	"github.com/stevenxie/api"
)

// A MoodFetcher can fetch moods from a MoodSource, and save them into a
// MoodService.
type MoodFetcher struct {
	src    api.MoodSource
	svc    api.MoodService
	logger *zap.SugaredLogger
}

// NewMoodFetcher builds a MoodFetcher.
func NewMoodFetcher(src api.MoodSource, svc api.MoodService) *MoodFetcher {
	return &MoodFetcher{
		src:    src,
		svc:    svc,
		logger: util.NoopLogger,
	}
}

// SetLogger sets the logger that MoodFetcher uses.
func (mf *MoodFetcher) SetLogger(logger *zap.SugaredLogger) {
	if logger == nil {
		logger = util.NoopLogger
	}
	mf.logger = logger
}

// FetchMoods fetches new moods from mf.Src, and places them in mf.Svc.
func (mf *MoodFetcher) FetchMoods() error {
	mf.l().Info("Fetching new moods...")

	// Fetch new moods from source.
	nmoods, err := mf.src.GetNewMoods()
	if err != nil {
		mf.l().Errorf("Error getting new moods from source: %v", err)
		return errors.Errorf("job: getting new moods from source: %w", err)
	}
	mf.l().Debugf("Got new moods from source: %+v", nmoods)

	// Get last saved mood.
	lmoods, err := mf.svc.ListMoods(1, 0)
	if err != nil {
		mf.l().Errorf("Error getting last moods from service: %v", err)
		return errors.Errorf("job: getting last moods from service: %w", err)
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
		mf.l().Infof("No new moods to save.")
		return nil
	}

	// Save new moods to service.
	if err = mf.svc.CreateMoods(keep); err != nil {
		mf.l().Errorf("Error while creating moods in service: %v", err)
		return errors.Errorf("job: creating moods in service: %w", err)
	}
	mf.l().Infof("Saved %d new moods.", len(keep))
	return nil
}

func (mf *MoodFetcher) l() *zap.SugaredLogger { return mf.logger }
