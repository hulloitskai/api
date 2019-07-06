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
	// A LocationService implements a latest-coordinates-preloading
	// api.LocationService using a geo.Geocoder and a Coordinator.
	LocationService struct {
		streamer *PollStreamer
		geocoder geo.Geocoder
		log      *logrus.Logger

		mux      sync.Mutex
		lastSeen *geo.Coordinate
		err      error
	}

	// A LocationHistorian can get the latest coordinates for a particular person.
	LocationHistorian interface {
		LatestCoordinates() (*geo.Coordinate, error)
	}

	// An LSOption configures a LocationPreloader.
	LSOption func(*LocationService)
)

var _ api.LocationService = (*LocationService)(nil)

// NewLocationPreloader creates a new LocationPreloader.
func NewLocationPreloader(
	historian LocationHistorian,
	geo geo.Geocoder,
	interval time.Duration,
	opts ...LSOption,
) *LocationService {
	ls := &LocationService{
		geocoder: geo,
		log:      zero.Logger(),
	}
	for _, opt := range opts {
		opt(ls)
	}

	// Configure streamer.
	action := func() (zero.Interface, error) {
		return historian.LatestCoordinates()
	}
	ls.streamer = NewPollStreamer(action, interval)

	go ls.populateCache()
	return ls
}

// WithLSLogger configures a LocationPreloader's logger.
func WithLSLogger(log *logrus.Logger) LSOption {
	return func(lp *LocationService) { lp.log = log }
}

func (ls *LocationService) populateCache() {
	for result := range ls.streamer.Stream() {
		var (
			lastSeen *geo.Coordinate
			err      error
		)

		switch v := result.(type) {
		case error:
			err = v
			ls.log.WithError(err).Error("Failed to load last seen position.")
		case *geo.Coordinate:
			lastSeen = v
		default:
			ls.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Errorf("stream: unexpected value '%s' from upstream")
		}

		ls.mux.Lock()
		ls.lastSeen = lastSeen
		ls.err = err
		ls.mux.Unlock()
	}
}

// Stop stops the LocationPreloader.
func (ls *LocationService) Stop() { ls.streamer.Stop() }

// LastSeen returns the authenticated user's last seen location.
func (ls *LocationService) LastSeen() (*geo.Coordinate, error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()
	return ls.lastSeen, ls.err
}

// CurrentCity returns the authenticated user's current city.
func (ls *LocationService) CurrentCity() (city string, err error) {
	coord, err := ls.LastSeen()
	if err != nil {
		return "", errors.Errorf("stream: determining last seen position: %w", err)
	}
	if coord == nil {
		return "", errors.New("stream: no position data available")
	}
	return geo.CityAt(ls.geocoder, *coord)
}

// CurrentRegion returns the authenticated user's current region.
func (ls *LocationService) CurrentRegion() (*geo.Location, error) {
	coord, err := ls.LastSeen()
	if err != nil {
		return nil, errors.Errorf("stream: determining last seen position: %w", err)
	}
	if coord == nil {
		return nil, errors.New("stream: no position data available")
	}
	return geo.RegionAt(ls.geocoder, *coord)
}
