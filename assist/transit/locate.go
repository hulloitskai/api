package transit

import (
	"context"

	"go.stevenxie.me/api/location"
)

type (
	// A Locator can locate nearby departures.
	Locator interface {
		NearbyDepartures(
			ctx context.Context,
			pos location.Coordinates,
			cfg NearbyDeparturesConfig,
		) ([]NearbyDeparture, error)
	}

	// A NearbyDeparture is a Departure that is occurring nearby.
	NearbyDeparture struct {
		Departure `json:"departure"`

		// How far away the departure is, in meters.
		Distance int `json:"distance"`
	}
)

type (
	// A LocatorService wraps a Locator with a friendlier API and logging.
	LocatorService interface {
		NearbyDepartures(
			ctx context.Context,
			pos location.Coordinates,
			opts ...NearbyDeparturesOption,
		) ([]NearbyDeparture, error)
	}

	// A NearbyDeparturesConfig contains optional parameters for
	// Locator.NearbyDepartures.
	NearbyDeparturesConfig struct {
		Radius        *uint // the search radius, in meters
		MaxStations   *uint // max number of stations to look up
		MaxPerStation *uint // max number of departures per station
	}

	// A NearbyDeparturesOption modifies a NearbyDepartureConfig.
	NearbyDeparturesOption func(*NearbyDeparturesConfig)
)
