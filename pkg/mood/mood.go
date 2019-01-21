package mood

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Mood describes the record of a mood.
type Mood struct {
	gorm.Model `json:"-"`
	ExtID      string    `json:"extId"`
	Mood       string    `json:"mood"`
	Valence    int       `json:"valence"`
	Context    string    `json:"context"`
	Reason     string    `json:"reason"`
	Date       time.Time `json:"date"`
}
