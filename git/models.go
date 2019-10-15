package git

import (
	"context"
	"time"

	"github.com/google/go-github/v25/github"
)

type (
	// A Commit represents a Git commit.
	Commit struct {
		SHA       string        `json:"sha"`
		Author    CommitAuthor  `json:"author"`
		Committer *CommitAuthor `json:"committer,omitempty"`
		Message   string        `json:"message"`
		URL       string        `json:"url"`
		Repo      Repo          `json:"repo"`
		Timestamp time.Time     `json:"timestamp"`
	}

	// A CommitAuthor forwards a definition.
	CommitAuthor github.CommitAuthor
)

// A Repo represents a Git repository.
type Repo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// A Source can retrieve recent Git commits from various projects.
type Source interface {
	RecentCommits(ctx context.Context, limit int) ([]*Commit, error)
}
