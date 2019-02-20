package api

import (
	"fmt"
	"time"
)

// Mood describes the record of a mood.
type Mood struct {
	ID        string    `json:"id"`
	ExtID     int64     `json:"extId"`
	Moods     []string  `json:"moods"`
	Valence   int       `json:"valence"`
	Context   []string  `json:"context"`
	Reason    string    `json:"-"`
	Timestamp time.Time `json:"timestamp"`
}

func (m *Mood) String() string { return fmt.Sprintf("%+v", *m) }

// MoodService is capable of storing and retrieving Moods.
type MoodService interface {
	GetMood(id string) (*Mood, error)
	ListMoods(limit, offset int) ([]*Mood, error)
	CreateMood(mood *Mood) error
	CreateMoods(moods []*Mood) error
}

// A MoodSource can be polled for new moods.
type MoodSource interface{ GetNewMoods() ([]*Mood, error) }
