package git // import "go.stevenxie.me/api/git"

import "context"

type (
	// A Service handles requests for my recent commits information.
	Service interface {
		RecentCommits(
			ctx context.Context,
			opts ...RecentCommitsOption,
		) ([]*Commit, error)
	}

	// A RecentCommitsConfig configures a Service.RecentCommits method.
	RecentCommitsConfig struct {
		Limit int
	}

	// A RecentCommitsOption modifies a RecentCommitsConfig.
	RecentCommitsOption func(*RecentCommitsConfig)
)
