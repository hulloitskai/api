package gitgql

import (
	"context"

	"go.stevenxie.me/api/v2/git"
)

// NewQuery creates a new Query.
func NewQuery(svc git.Service) Query {
	return Query{svc: svc}
}

// A Query resolves queries for my scheduling-related data.
type Query struct {
	svc git.Service
}

// RecentCommits looks up my recent commits.
func (q Query) RecentCommits(
	ctx context.Context,
	limit *int,
) ([]git.Commit, error) {
	return q.svc.RecentCommits(
		ctx,
		func(opt *git.RecentCommitsOptions) {
			if limit != nil {
				opt.Limit = *limit
			}
		},
	)
}
