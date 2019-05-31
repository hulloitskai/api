package github

import (
	"context"

	errors "golang.org/x/xerrors"

	"github.com/google/go-github/v25/github"
	"github.com/stevenxie/api/pkg/git"
)

const (
	eventTypePush = "PushEvent"
	maxEventsPage = 10
)

// RecentCommits retrieves the latest `limit` Git commits across unique
// repositories.
func (c *Client) RecentCommits(limit int) ([]*git.Commit, error) {
	return c.recentCommits(
		limit,
		0,
		make([]*git.Commit, 0, limit),
		make(map[string]struct{}),
	)
}

func (c *Client) recentCommits(
	limit, page int,
	commits []*git.Commit,
	seenRepos map[string]struct{},
) ([]*git.Commit, error) {
	// Get current user.
	login, err := c.CurrentUserLogin()
	if err != nil {
		return nil, errors.Errorf("github: getting current user: %w", err)
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
			return nil, errors.Errorf("github: failed to parse event payload: %w",
				err)
		}
		pushPayload := payload.(*github.PushEvent)
		if len(pushPayload.Commits) == 0 {
			continue
		}

		commit := pushPayload.Commits[0]
		var (
			repo = e.GetRepo()
			cm   = &git.Commit{
				SHA:       commit.GetSHA(),
				Author:    commit.GetAuthor(),
				Committer: commit.GetCommitter(),
				Message:   commit.GetMessage(),
				URL:       commit.GetURL(),
				Repo: &git.Repo{
					Name: repo.GetName(),
					URL:  repo.GetURL(),
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
		seenRepos[repo.GetName()] = struct{}{}
	}

	if (len(commits) < limit) && (page < maxEventsPage) {
		return c.recentCommits(limit, page+1, commits, seenRepos)
	}
	return commits, nil
}
