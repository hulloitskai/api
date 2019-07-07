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

		mux     sync.Mutex
		segment *geo.Segment
		err     error
	}

	// An LSOption configures a LocationPreloader.
	LSOption func(*LocationService)

	// A RecentLocationsService can fetch data relating to one's recent locations.
	RecentLocationsService interface{ LastSegment() (*geo.Segment, error) }
)

var _ api.LocationService = (*LocationService)(nil)

// NewLocationPreloader creates a new LocationPreloader.
func NewLocationPreloader(
	locations RecentLocationsService,
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
	action := func() (zero.Interface, error) { return locations.LastSegment() }
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
			segment *geo.Segment
			err     error
		)

		switch v := result.(type) {
		case error:
			err = v
			ls.log.WithError(err).Error("Failed to load last seen position.")
		case *geo.Segment:
			segment = v
		default:
			ls.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Errorf("stream: unexpected value '%s' from upstream")
		}

		ls.mux.Lock()
		ls.segment = segment
		ls.err = err
		ls.mux.Unlock()
	}
}

// Stop stops the LocationPreloader.
func (ls *LocationService) Stop() { ls.streamer.Stop() }

// LastSegment returns the authenticated user's latest location history segment.
func (ls *LocationService) LastSegment() (*geo.Segment, error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()
	copy := *ls.segment
	return &copy, nil
}

// LastPosition returns the authenticated user's last known position.
func (ls *LocationService) LastPosition() (*geo.Coordinate, error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()
	if ls.segment == nil {
		return nil, nil
	}
	coords := ls.segment.Coordinates
	if len(coords) == 0 {
		return nil, nil
	}
	copy := coords[len(coords)-1]
	return &copy, nil
}

// CurrentCity returns the authenticated user's current city.
func (ls *LocationService) CurrentCity() (city string, err error) {
	coord, err := ls.LastPosition()
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
	coord, err := ls.LastPosition()
	if err != nil {
		return nil, errors.Errorf("stream: determining last seen position: %w", err)
	}
	if coord == nil {
		return nil, errors.New("stream: no position data available")
	}
	return geo.RegionAt(ls.geocoder, *coord)
}
