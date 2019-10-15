package music

import "context"

// A Source can retrieve music-related values.
type Source interface {
	GetTrack(ctx context.Context, id string) (*Track, error)

	GetAlbumTracks(
		ctx context.Context,
		id string,
		limit, offset int,
	) ([]Track, error)

	GetArtistAlbums(
		ctx context.Context,
		id string,
		limit, offset int,
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

	// PaginationConfig configures a paginated request.
	PaginationConfig struct {
		Limit  int
		Offset int
	}

	// PaginationOption modifies a PaginationConfig.
	PaginationOption func(*PaginationConfig)
)
