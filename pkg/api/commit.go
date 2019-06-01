package api

import (
	"time"

	"github.com/google/go-github/v25/github"
)

// A GitCommit represents a Git commit.
type GitCommit struct {
	SHA       string               `json:"sha"`
	Author    *github.CommitAuthor `json:"author"`
	Committer *github.CommitAuthor `json:"committer,omitempty"`
	Message   string               `json:"message"`
	URL       string               `json:"url"`
	Repo      *GitRepo             `json:"repo"`
	Timestamp time.Time            `json:"timestamp"`
}

// A GitRepo represents a Git repository.
type GitRepo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// A GitCommitsService can retrieve recent commits from various projects.
type GitCommitsService interface {
	RecentGitCommits(limit int) ([]*GitCommit, error)
}
