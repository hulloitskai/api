package mood

import (
	"time"
)

// Mood describes the record of a mood.
type Mood struct {
	ID        string    `json:"id"        bson:"-"`
	ExtID     int64     `json:"extId"     bson:"extId"`
	Moods     []string  `json:"moods"`
	Valence   int       `json:"valence"`
	Context   []string  `json:"context"`
	Reason    string    `json:"-"`
	Timestamp time.Time `json:"timestamp"`
}

// Source is a source of Mood objects. Moods originate from a Source.
type Source interface {
	// FetchMoods fetches the last `limit` Moods.
	FetchMoods(limit int) ([]*Mood, error)
}

// Repo is a repository of Mood objecs. Moods are stored (persisted) in a Repo.
type Repo interface {
	SelectMoods(limit int, startID string) ([]*Mood, error)
	InsertMoods(moods []*Mood) error
	// UpdateMood(mood *Mood) error
	// DeleteMood(mood *Mood) error
}
