package git // import "go.stevenxie.me/api/v2/git"

import "context"

type (
	// A Service handles requests for my recent commits information.
	Service interface {
		RecentCommits(
			ctx context.Context,
			opts ...RecentCommitsOption,
		) ([]Commit, error)
	}

	// RecentCommitsOptions are option parameters for Service.RecentCommits.
	RecentCommitsOptions struct {
		Limit int
	}

	// A RecentCommitsOption modifies a RecentCommitsOptions.
	RecentCommitsOption func(*RecentCommitsOptions)
)
