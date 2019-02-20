package airtable

import (
	errors "golang.org/x/xerrors"
)

// A Provider provides data using Airtable as an underlying datasource.
type Provider struct {
	client *Client

	moodSource *MoodSource
}

// NewProvider creates a new Provider.
func NewProvider(apiKey, baseID string) (*Provider, error) {
	client, err := NewClient(apiKey, baseID)
	if err != nil {
		return nil, errors.Errorf("airtable: creating client: %w", err)
	}
	return &Provider{
		client:     client,
		moodSource: newMoodSource(client),
	}, nil
}

// MoodSource returns an implementation of api.MoodSource that uses Airtable.
func (p *Provider) MoodSource() *MoodSource { return p.moodSource }
