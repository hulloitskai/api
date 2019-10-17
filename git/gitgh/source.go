package gitgh

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	ghlib "github.com/google/go-github/v25/github"
	"go.stevenxie.me/gopkg/zero"

	"go.stevenxie.me/api/git"
	"go.stevenxie.me/api/pkg/github"
)

const _homeURL = "https://github.com"

// NewSource creates a new CommitsService.
func NewSource(c *github.Client) git.Source {
	return source{client: c}
}

const (
	_pushEventType = "PushEvent"
	_maxEventsPage = 10
)

type source struct {
	client *github.Client
}

var _ git.Source = (*source)(nil)

// RecentCommits retrieves the latest `limit` Git commits across unique
// repositories.
func (svc source) RecentCommits(
	ctx context.Context,
	limit int,
) ([]git.Commit, error) {
	// Get current user login.
	login, err := svc.client.CurrentUserLogin()
	if err != nil {
		return nil, errors.Wrap(err, "gitstatgh: getting current user")
	}

	var (
		cms       = make([]git.Commit, 0, limit)
		seenRepos = make(map[string]zero.Struct)
	)

	// Loop through all event pages, looking for git.
PageLoop:
	for page := 0; page < _maxEventsPage; page++ {
		// List user events.
		events, _, err := svc.client.GitHub().Activity.ListEventsPerformedByUser(
			ctx,
			login, false,
			&ghlib.ListOptions{Page: page},
		)
		if err != nil {
			return nil, err
		}

		// Filter and parse events.
		for _, e := range events {
			if e.GetType() != _pushEventType {
				continue
			}
			if _, ok := seenRepos[e.GetRepo().GetName()]; ok { // enforce uniqueness
				continue
			}

			// Extract latest commit from push event.
			var pushCommit ghlib.PushEventCommit
			{
				payload, err := e.ParsePayload()
				if err != nil {
					return nil, errors.Wrap(
						err,
						"gitstatgh: failed to parse event payload",
					)
				}
				push := payload.(*ghlib.PushEvent)
				if len(push.Commits) == 0 {
					continue
				}
				pushCommit = push.Commits[0]
			}

			author := git.CommitAuthor(*pushCommit.GetAuthor())

			var committer *git.CommitAuthor
			if c := pushCommit.GetCommitter(); c != nil {
				ca := git.CommitAuthor(*c)
				committer = &ca
			}

			// Construct git.Commit.
			var (
				repo   = e.GetRepo().GetName()
				commit = git.Commit{
					SHA:       pushCommit.GetSHA(),
					Author:    author,
					Committer: committer,
					Message:   pushCommit.GetMessage(),
					URL: fmt.Sprintf(
						"%s/%s/commit/%s",
						_homeURL,
						repo,
						pushCommit.GetSHA(),
					),
					Repo: git.Repo{
						Name: repo,
						URL:  fmt.Sprintf("%s/%s", _homeURL, repo),
					},
					Timestamp: e.GetCreatedAt(),
				}
			)

			// Append to cms, add repo to seen set.
			cms = append(cms, commit)
			if len(cms) == limit {
				break PageLoop
			}

			seenRepos[repo] = zero.Empty()
		}
	}

	return cms, nil
}

func committerFromGH(ca *ghlib.CommitAuthor) *git.CommitAuthor {
	if ca == nil {
		return nil
	}
	conv := git.CommitAuthor(*ca)
	return &conv
}
