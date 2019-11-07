package spotify

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/openlyinc/pointy"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"go.stevenxie.me/api/v2/music"
	"go.stevenxie.me/api/v2/pkg/basic"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"
)

// NewController creates a new music.Controller.
func NewController(c *spotify.Client, opts ...basic.Option) music.Controller {
	opt := basic.BuildOptions(opts...)
	return controller{
		client: c,
		log:    logutil.WithComponent(opt.Logger, (*controller)(nil)),
		tracer: opt.Tracer,
	}
}

type controller struct {
	client *spotify.Client
	log    *logrus.Entry
	tracer opentracing.Tracer
}

var _ music.Controller = (*controller)(nil)

func (ctrl controller) Play(ctx context.Context, s *music.Selector) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, ctrl.tracer,
		name.OfFunc(controller.Play),
	)
	defer span.Finish()

	log := logutil.
		WithMethod(ctrl.log, controller.Play).
		WithContext(ctx)

	// Derive Spotify URI.
	var uri *string
	if s != nil {
		log := log.WithField("selector", *s)
		if err := s.Validate(); err != nil {
			log.WithError(err).Error("Invalid music.Cond.")
			return errors.Wrap(err, "spotify: validate music.Cond")
		}
		if u := s.URI; u != nil {
			uri = u
		} else {
			resources := []struct {
				Kind  string
				Value *music.Resource
			}{
				{Kind: "track", Value: s.Track},
				{Kind: "album", Value: s.Album},
				{Kind: "artist", Value: s.Artist},
				{Kind: "playlist", Value: s.Playlist},
			}
			for _, r := range resources {
				if v := r.Value; v != nil {
					uri = pointy.String(fmt.Sprintf("spotify:%s:%s", r.Kind, v.ID))
					break
				}
			}
		}
		log.WithField("uri", *uri).Trace("Derived resource URI.")
	}

	// Build play options, and execute.
	var opts spotify.PlayOptions
	if uri != nil {
		var (
			u = *uri
			l = len("spotify:")
		)
		kind := u[l : l+strings.IndexByte(u[l:], ':')]
		if kind == "track" {
			opts.URIs = []spotify.URI{spotify.URI(u)}
		} else {
			su := spotify.URI(u)
			opts.PlaybackContext = &su
		}
		log.Trace("Playing specified resource...")
	} else {
		log.Trace("Resuming current track.")
	}
	return errors.WithMessage(ctrl.client.PlayOpt(&opts), "spotify")
}

func (ctrl controller) Pause(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, ctrl.tracer,
		name.OfFunc(controller.Pause),
	)
	defer span.Finish()

	return errors.WithMessage(ctrl.client.Pause(), "spotify")
}
