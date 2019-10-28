package about // import "go.stevenxie.me/api/v2/about"

import "context"

// A Service can get personal information.
type Service interface {
	GetAbout(ctx context.Context) (*About, error)
	GetMasked(ctx context.Context) (*Masked, error)
}
