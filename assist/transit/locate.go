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
			opts ...NearbyDeparturesOption,
		) ([]NearbyDeparture, error)
	}

	// A NearbyDeparture is a Departure that is occurring nearby.
	NearbyDeparture struct {
		Departure `json:"departure"`

		// How far away the departure is, in meters.
		Distance int `json:"distance"`
	}

	// A NearbyDeparturesConfig contains optional parameters for
	// Locator.NearbyDepartures.
	NearbyDeparturesConfig struct {
		Radius        *uint // the search radius, in meters
		MaxStations   *uint
		MaxPerStation *uint
	}

	// A NearbyDeparturesOption modifies a NearbyDepartureConfig.
	NearbyDeparturesOption func(*NearbyDeparturesConfig)
)
