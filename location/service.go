package location // import "go.stevenxie.me/api/v2/location"

import (
	"context"
	"time"
)

// A Service provides information about my recent locations.
type Service interface {
	PositionService
	HistoryService
}

type (
	// A PositionService can determine my current position.
	PositionService interface {
		TimeZoneService

		CurrentPosition(ctx context.Context) (*Coordinates, error)
		CurrentCity(ctx context.Context) (string, error)
		CurrentRegion(
			ctx context.Context,
			opts ...CurrentRegionOption,
		) (*Place, error)
	}

	// CurrentRegionOptions are option parameters for
	// PostionService.CurrentRegion.
	CurrentRegionOptions struct {
		IncludeTimeZone bool
	}

	// A CurrentRegionOption modifies a CurrentRegionOptions.
	CurrentRegionOption func(*CurrentRegionOptions)
)

// A TimeZoneService can get my current time zone.
type TimeZoneService interface {
	CurrentTimeZone(ctx context.Context) (*time.Location, error)
}

// A HistoryService can determine where I've been in the past.
type HistoryService interface {
	RecentHistory(ctx context.Context) ([]HistorySegment, error)
	GetHistory(ctx context.Context, date time.Time) ([]HistorySegment, error)
}
