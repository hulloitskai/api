package data

import (
	"errors"

	"github.com/stevenxie/api/pkg/mood"
	"github.com/stevenxie/api/pkg/mood/adapter"
	ess "github.com/unixpickle/essentials"
)

// RepoSet is a set of data repositories.
type RepoSet struct {
	MoodRepo mood.Repo
}

// NewRepoSet constructs a RepoSet from a DriverSet.
func NewRepoSet(drivers *DriverSet) (*RepoSet, error) {
	if drivers == nil {
		panic(errors.New("data: cannot make RepoSet wth nil db"))
	}

	moodRepo, err := adapter.NewMongoAdapter(drivers.Mongo)
	if err != nil {
		return nil, ess.AddCtx("data: creating mood repo", err)
	}
	return &RepoSet{
		MoodRepo: moodRepo,
	}, nil
}
