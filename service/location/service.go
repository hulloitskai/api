package location

type (
	// A Service provides information about my recent locations.
	Service interface {
		CurrentCity() (city string, err error)
		CurrentRegion() (*Place, error)
		LastPosition() (*Coordinates, error)
		RecentHistory() ([]*HistorySegment, error)
	}

	// A AccessService can validate location access codes.
	AccessService interface {
		IsValidCode(code string) (valid bool, err error)
	}
)
