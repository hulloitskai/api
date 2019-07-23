package github

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v25/github"

	gh "go.stevenxie.me/api/pkg/github"
	"go.stevenxie.me/api/pkg/zero"
	cm "go.stevenxie.me/api/service/commits"
)

const (
	eventTypePush = "PushEvent"
	maxEventsPage = 10
)

// A CommitsService implements a commits.Service using an authenticated
// http.Client.
type CommitsService struct{ client *gh.Client }

// NewCommitsService creates a new CommitsService.
func NewCommitsService(c *gh.Client) CommitsService {
	return CommitsService{client: c}
}

var _ cm.Service = (*CommitsService)(nil)

// RecentCommits retrieves the latest `limit` Git commits across unique
// repositories.
func (svc CommitsService) RecentCommits(limit int) (cm.Commits, error) {
	return svc.recentCommits(
		limit,
		0,
		make(cm.Commits, 0, limit),
		make(map[string]zero.Struct),
	)
}

func (svc CommitsService) recentCommits(
	limit, page int,
	commits cm.Commits,
	seenRepos map[string]zero.Struct,
) (cm.Commits, error) {
	// Get current user.
	login, err := svc.client.CurrentUserLogin()
	if err != nil {
		return nil, errors.Wrap(err, "github: getting current user")
	}

	// List user events.
	events, _, err := svc.client.GHClient().Activity.ListEventsPerformedByUser(
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
		paycomm := pushPayload.Commits[0]

		var (
			repo    = e.GetRepo()
			baseURL = svc.client.BaseURL()
			commit  = &cm.Commit{
				SHA:       paycomm.GetSHA(),
				Author:    paycomm.GetAuthor(),
				Committer: paycomm.GetCommitter(),
				Message:   paycomm.GetMessage(),
				URL: fmt.Sprintf(
					"%s/%s/commit/%s",
					baseURL,
					repo.GetName(),
					paycomm.GetSHA(),
				),
				Repo: &cm.Repo{
					Name: repo.GetName(),
					URL:  fmt.Sprintf("%s/%s", baseURL, repo.GetName()),
				},
				Timestamp: e.GetCreatedAt(),
			}
		)

		// Append to commits.
		commits = append(commits, commit)
		if len(commits) == limit {
			break
		}

		// Add repo to seen cache.
		seenRepos[repo.GetName()] = zero.Empty()
	}

	if (len(commits) < limit) && (page < maxEventsPage) {
		return svc.recentCommits(limit, page+1, commits, seenRepos)
	}
	return commits, nil
}
