package airtable

import "time"

// A Mood is an Airtable record from the 'moods' table.
type Mood struct {
	ID        int64     `mapstructure:"id"`
	Moods     []string  `mapstructure:"moods"`
	Valence   int       `mapstructure:"valence"`
	Context   []string  `mapstructure:"context"`
	Reason    string    `mapstructure:"reason"`
	Timestamp time.Time `mapstructure:"timestamp"`
}

// Moods retrieves the last `limit` moods from Airtable.
func (c *Client) Moods(limit int) ([]*Mood, error) {
	var (
		opts = fetchOpts{
			Limit: limit,
			Sort: []sortConfig{{
				Field:     "id",
				Direction: "desc",
			}},
		}
		moods []*Mood
	)
	err := c.fetchRecords(c.cfg.MoodTableName, &moods, &opts)
	return moods, err
}
