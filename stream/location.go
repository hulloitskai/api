package stream

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	errors "golang.org/x/xerrors"

	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/geo"
	"github.com/stevenxie/api/pkg/zero"
)

type (
	// A LocationPreloader preloads last seen position data for a LocationService.
	LocationPreloader struct {
		streamer *PollStreamer
		geocoder geo.Geocoder
		log      *logrus.Logger

		mux      sync.Mutex
		lastSeen *geo.Coordinate
		err      error
	}

	// An LPOption configures a LocationPreloader.
	LPOption func(*LocationPreloader)
)

var _ api.LocationService = (*LocationPreloader)(nil)

// NewLocationPreloader creates a new LocationPreloader.
func NewLocationPreloader(
	svc api.LocationService,
	geo geo.Geocoder,
	interval time.Duration,
	opts ...LPOption,
) *LocationPreloader {
	lp := &LocationPreloader{
		geocoder: geo,
		log:      zero.Logger(),
	}
	for _, opt := range opts {
		opt(lp)
	}

	// Configure streamer.
	action := func() (zero.Interface, error) { return svc.LastSeen() }
	lp.streamer = NewPollStreamer(action, interval)

	go lp.populateCache()
	return lp
}

// WithLPLogger configures a LocationPreloader's logger.
func WithLPLogger(log *logrus.Logger) LPOption {
	return func(lp *LocationPreloader) { lp.log = log }
}

func (lp *LocationPreloader) populateCache() {
	for result := range lp.streamer.Stream() {
		var (
			lastSeen *geo.Coordinate
			err      error
		)

		switch v := result.(type) {
		case error:
			err = v
			lp.log.WithError(err).Error("Failed to load last seen position.")
		case *geo.Coordinate:
			lastSeen = v
		}

		lp.mux.Lock()
		lp.lastSeen = lastSeen
		lp.err = err
		lp.mux.Unlock()
	}
}

// LastSeen returns the authenticated user's last seen location.
func (lp *LocationPreloader) LastSeen() (*geo.Coordinate, error) {
	lp.mux.Lock()
	defer lp.mux.Unlock()
	return lp.lastSeen, lp.err
}

// CurrentCity returns the authenticated user's current city.
func (lp *LocationPreloader) CurrentCity() (city string, err error) {
	coord, err := lp.LastSeen()
	if err != nil {
		return "", errors.Errorf("stream: determining last seen position: %w", err)
	}
	if coord == nil {
		return "", errors.New("stream: no position data available")
	}
	return geo.CityAt(lp.geocoder, *coord)
}

// Stop stops the LocationPreloader.
func (lp *LocationPreloader) Stop() { lp.streamer.Stop() }
