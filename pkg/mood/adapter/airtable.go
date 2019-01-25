package adapter

import (
	"github.com/stevenxie/api/pkg/data/airtable"
	"github.com/stevenxie/api/pkg/mood"
)

// An AirtableAdapter implements moods.Source for an airtable.Client.
type AirtableAdapter struct {
	*airtable.Client
}

// FetchMoods implements moods.Source.
func (aa AirtableAdapter) FetchMoods(limit int) ([]*mood.Mood, error) {
	results, err := aa.Client.Moods(limit)
	if err != nil {
		return nil, err
	}

	moods := make([]*mood.Mood, len(results))
	for i, result := range results {
		moods[i] = &mood.Mood{
			ExtID:     result.ID,
			Moods:     result.Moods,
			Valence:   result.Valence,
			Context:   result.Context,
			Reason:    result.Reason,
			Timestamp: result.Timestamp,
		}
	}
	return moods, nil
}
