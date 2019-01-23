package airtable

import "time"

// A Mood is an Airtable record from the 'moods' table.
type Mood struct {
	ID      int64     `mapstructure:"id"`
	Mood    []string  `mapstructure:"mood"`
	Valence int       `mapstructure:"valence"`
	Context []string  `mapstructure:"context"`
	Reason  string    `mapstructure:"reason"`
	Date    time.Time `mapstructure:"date"`
}

// Moods retrieves the last `limit` moods from Airtable.
func (c *Client) Moods(limit int) ([]*Mood, error) {
	var moods []*Mood
	if err := c.UnmarshalRecords(c.cfg.MoodTableName, c.cfg.MoodTableView, limit,
		&moods); err != nil {
		return nil, err
	}
	return moods, nil
}
