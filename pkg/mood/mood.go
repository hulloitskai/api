package mood

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Mood describes the record of a mood.
type Mood struct {
	gorm.Model `json:"-"`
	ExtID      int64     `json:"extId"`
	Moods      []string  `json:"moods"`
	Valence    int       `json:"valence"`
	Context    []string  `json:"context"`
	Reason     string    `json:"-"`
	Timestamp  time.Time `json:"timestamp"`
}

// Source is a source of Mood objects. Moods originate from a Source.
type Source interface {
	// FetchMoods fetches the last `limit` Moods.
	FetchMoods(limit int) ([]*Mood, error)
}
