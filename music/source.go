package music

import "context"

// A Source can retrieve music-related values.
type Source interface {
	GetTrack(ctx context.Context, id string) (*Track, error)

	GetAlbumTracks(
		ctx context.Context,
		id string,
		opt PaginationOptions,
	) ([]Track, error)

	GetArtistAlbums(
		ctx context.Context,
		id string,
		opt PaginationOptions,
	) ([]Album, error)
}

type (
	// A SourceService wraps a Source with a more user-friendly API.
	SourceService interface {
		GetTrack(ctx context.Context, id string) (*Track, error)

		GetAlbumTracks(
			ctx context.Context,
			id string,
			opts ...PaginationOption,
		) ([]Track, error)

		GetArtistAlbums(
			ctx context.Context,
			id string,
			opts ...PaginationOption,
		) ([]Album, error)
	}

	// PaginationOptions are option parameters for a paginated request.
	PaginationOptions struct {
		Limit  int
		Offset int
	}

	// PaginationOption modifies a PaginationOptions.
	PaginationOption func(*PaginationOptions)
)
