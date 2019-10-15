package musicsvc

import "go.stevenxie.me/api/music"

// NewService creates a new music.Service.
func NewService(
	src music.SourceService,
	curr music.CurrentService,
	ctrl music.ControlService,
) music.Service {
	return service{
		SourceService:  src,
		CurrentService: curr,
		ControlService: ctrl,
	}
}

type service struct {
	music.SourceService
	music.CurrentService
	music.ControlService
}

var _ music.Service = (*service)(nil)
