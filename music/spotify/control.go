package spotify

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/zmb3/spotify"
	"go.stevenxie.me/api/v2/music"
)

// NewController creates a new music.Controller.
func NewController(c *spotify.Client) music.Controller {
	return controller{client: c}
}

type controller struct {
	client *spotify.Client
}

var _ music.Controller = (*controller)(nil)

func (ctrl controller) Play(_ context.Context, uri *string) error {
	var opts spotify.PlayOptions
	if uri != nil {
		opts.URIs = []spotify.URI{spotify.URI(*uri)}
	}
	return errors.WithMessage(ctrl.client.PlayOpt(&opts), "spotify")
}

func (ctrl controller) Pause(_ context.Context) error {
	return errors.WithMessage(ctrl.client.Pause(), "spotify")
}
