package commits

import (
	"time"

	"github.com/google/go-github/v25/github"
)

type (
	// A Commit represents a Git commit.
	Commit struct {
		SHA       string               `json:"sha"`
		Author    *github.CommitAuthor `json:"author"`
		Committer *github.CommitAuthor `json:"committer,omitempty"`
		Message   string               `json:"message"`
		URL       string               `json:"url"`
		Repo      *Repo                `json:"repo"`
		Timestamp time.Time            `json:"timestamp"`
	}

	// Commits are a set of Commits.
	Commits []*Commit

	// A Repo represents a Git repository.
	Repo struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
)

// A Service can retrieve recent Git commits from various projects.
type Service interface {
	RecentCommits(limit int) (Commits, error)
}
