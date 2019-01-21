package airtable

import (
	"time"

	at "github.com/fabioberger/airtable-go"
	m "github.com/stevenxie/api/pkg/mood"
	ess "github.com/unixpickle/essentials"
)

type mood struct {
	ID      string    `json:"id"`
	Mood    string    `json:"mood"`
	Valence int       `json:"valence"`
	Context string    `json:"context"`
	Reason  string    `json:"reason"`
	Date    time.Time `json:"date"`
}

// Moods retrieves the last `limit` moods from Airtable.
func (c *Client) Moods(limit int) ([]*m.Mood, error) {
	var raw []mood
	params := at.ListParameters{
		Fields:     []string{"valence", "context", "id", "mood", "reason", "date"},
		MaxRecords: limit,
		View:       "Main View",
	}
	if err := c.c.ListRecords("moods", &raw, params); err != nil {
		return nil, ess.AddCtx("airtable", err)
	}

	moods := make([]*m.Mood, len(raw))
	for i := range raw {
		rm := &raw[i]
		moods[i] = &m.Mood{
			ExtID:   rm.ID,
			Mood:    rm.Mood,
			Valence: rm.Valence,
			Context: rm.Context,
			Reason:  rm.Reason,
			Date:    rm.Date,
		}
	}
	return moods, nil
}
