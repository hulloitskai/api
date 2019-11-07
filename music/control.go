package music

import (
	"context"
	"reflect"

	"github.com/cockroachdb/errors"
	validation "github.com/go-ozzo/ozzo-validation"
)

// A Controller can control my music player.
type Controller interface {
	Play(ctx context.Context, s *Selector) error
	Pause(ctx context.Context) error
}

// PlayResource configures the ControlService.Play method to play the resource
// specified by s.
func PlayResource(s Selector) PlayOption {
	return func(cfg *PlayOptions) { cfg.Selector = &s }
}

type (
	// A Selector is used to select a music resource.
	Selector struct {
		URI      *string   `json:"uri"`
		Track    *Resource `json:"track"`
		Album    *Resource `json:"album"`
		Artist   *Resource `json:"artist"`
		Playlist *Resource `json:"playlist"`
	}

	// A Resource specifies a music resource by its ID.
	Resource struct {
		ID string `json:"id"`
	}
)

var _ validation.Validatable = (*Selector)(nil)

// Validate returns an error if the Cond is not valid (i.e. it does not
// specify a valid resource).
func (s *Selector) Validate() error {
	var count int
	{
		v := reflect.ValueOf(s).Elem()
		for i := 0; i < v.NumField(); i++ {
			if !v.Field(i).IsNil() {
				count++
			}
		}
	}
	if count != 1 {
		return errors.New("music: one field must be specified")
	}
	return nil
}

type (
	// A ControlService wraps a Controller with a friendlier API.
	ControlService interface {
		Play(ctx context.Context, opts ...PlayOption) error
		Pause(ctx context.Context) error
	}

	// PlayOptions are option parameters for ControlService.Play.
	PlayOptions struct {
		Selector *Selector
	}

	// A PlayOption modifies a PlayOptions.
	PlayOption func(*PlayOptions)
)
