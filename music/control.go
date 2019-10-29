package music

import "context"

// A Controller can control my music player.
type Controller interface {
	Play(ctx context.Context, uri *string) error
	Pause(ctx context.Context) error
}

// PlayURI configures the ControlService.Play method to play the specified
// resource.
func PlayURI(uri string) PlayOption {
	return func(cfg *PlayOptions) { cfg.URI = &uri }
}

type (
	// A ControlService wraps a Controller with a friendlier API.
	ControlService interface {
		Play(ctx context.Context, opts ...PlayOption) error
		Pause(ctx context.Context) error
	}

	// PlayOptions are option parameters for ControlService.Play.
	PlayOptions struct {
		URI *string
	}

	// A PlayOption modifies a PlayOptions.
	PlayOption func(*PlayOptions)
)
