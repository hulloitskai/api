package github

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v25/github"
	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/zero"
)

const (
	baseURL       = "https://github.com"
	eventTypePush = "PushEvent"
	maxEventsPage = 10
)

// RecentGitCommits retrieves the latest `limit` Git commits across unique
// repositories.
func (c *Client) RecentGitCommits(limit int) ([]*api.GitCommit, error) {
	return c.recentCommits(
		limit,
		0,
		make([]*api.GitCommit, 0, limit),
		make(map[string]zero.Struct),
	)
}

func (c *Client) recentCommits(
	limit, page int,
	commits []*api.GitCommit,
	seenRepos map[string]zero.Struct,
) ([]*api.GitCommit, error) {
	// Get current user.
	login, err := c.CurrentUserLogin()
	if err != nil {
		return nil, errors.Wrap(err, "github: getting current user")
	}

	// List user events.
	events, _, err := c.ghc.Activity.ListEventsPerformedByUser(
		context.Background(),
		login,
		false,
		&github.ListOptions{Page: page},
	)
	if err != nil {
		return nil, err
	}

	// Filter and parse events.
	for _, e := range events {
		if e.GetType() != eventTypePush {
			continue
		}
		if _, ok := seenRepos[e.GetRepo().GetName()]; ok { // enforce uniqueness
			continue
		}

		payload, err := e.ParsePayload()
		if err != nil {
			return nil, errors.Wrap(err, "github: failed to parse event payload")
		}
		pushPayload := payload.(*github.PushEvent)
		if len(pushPayload.Commits) == 0 {
			continue
		}

		commit := pushPayload.Commits[0]
		var (
			repo = e.GetRepo()
			cm   = &api.GitCommit{
				SHA:       commit.GetSHA(),
				Author:    commit.GetAuthor(),
				Committer: commit.GetCommitter(),
				Message:   commit.GetMessage(),
				URL: fmt.Sprintf("%s/%s/commit/%s", baseURL, repo.GetName(),
					commit.GetSHA()),
				Repo: &api.GitRepo{
					Name: repo.GetName(),
					URL:  fmt.Sprintf("%s/%s", baseURL, repo.GetName()),
				},
				Timestamp: e.GetCreatedAt(),
			}
		)

		// Append to commits.
		commits = append(commits, cm)
		if len(commits) == limit {
			break
		}

		// Add repo to seen cache.
		seenRepos[repo.GetName()] = zero.Empty()
	}

	if (len(commits) < limit) && (page < maxEventsPage) {
		return c.recentCommits(limit, page+1, commits, seenRepos)
	}
	return commits, nil
}
