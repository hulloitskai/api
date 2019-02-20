package airtable

import (
	"time"

	"github.com/stevenxie/api"
)

// MoodsTable is the name of mood table in Airtable.
const MoodsTable = "moods"

// A MoodSource implements job.MoodSource for an Airtable client.
type MoodSource struct {
	client     *Client
	fetchLimit int // default 10
}

func newMoodSource(c *Client) *MoodSource {
	return &MoodSource{
		client:     c,
		fetchLimit: 10,
	}
}

// SetFetchLimit sets the fetch limit of a MoodSource, which is the number of
// moods that are retrieved each time.
func (ms *MoodSource) SetFetchLimit(limit int) { ms.fetchLimit = limit }

type moodRecord struct {
	ID        int64     `ms:"id"`
	Moods     []string  `ms:"moods"`
	Valence   int       `ms:"valence"`
	Context   []string  `ms:"context"`
	Reason    string    `ms:"reason"`
	Timestamp time.Time `ms:"timestamp"`
}

// GetNewMoods gets new moods from Airtable.
func (ms *MoodSource) GetNewMoods() ([]*api.Mood, error) {
	var (
		opts = FetchOptions{
			Limit: ms.fetchLimit,
			Sort: []SortConfig{{
				Field:     "id",
				Direction: "desc",
			}},
		}
		records []*moodRecord
	)
	if err := ms.client.FetchRecords(MoodsTable, &records, &opts); err != nil {
		return nil, err
	}

	// Unmarshal records to moods.
	moods := make([]*api.Mood, len(records))
	for i, record := range records {
		moods[i] = &api.Mood{
			ExtID:     record.ID,
			Moods:     record.Moods,
			Valence:   record.Valence,
			Context:   record.Context,
			Reason:    record.Reason,
			Timestamp: record.Timestamp,
		}
	}
	return moods, nil
}
