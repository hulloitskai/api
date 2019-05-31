package git

import (
	"time"

	"github.com/google/go-github/v25/github"
)

// A Repo represents a Git repository.
type Repo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// A Commit represents a Git commit.
type Commit struct {
	SHA       string               `json:"sha"`
	Author    *github.CommitAuthor `json:"author"`
	Committer *github.CommitAuthor `json:"committer,omitempty"`
	Message   string               `json:"message"`
	URL       string               `json:"url"`
	Repo      *Repo                `json:"repo"`
	Timestamp time.Time            `json:"timestamp"`
}

// A RecentCommitsService can retrieve recent commits from various projects.
type RecentCommitsService interface {
	RecentCommits(limit int) ([]*Commit, error)
}
