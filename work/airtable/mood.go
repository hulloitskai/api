package airtable

import (
	"time"

	"github.com/spf13/viper"
	"github.com/stevenxie/api/data/airtable"
	"github.com/stevenxie/api/internal/util"

	defaults "github.com/mcuadros/go-defaults"
	"github.com/stevenxie/api"
)

// MoodsTable is the name of mood table in Airtable.
const MoodsTable = "moods"

// MoodSourceConfig contains configuration details for a MoodSource.
type MoodSourceConfig struct {
	Limit int `ms:"limit" default:"10"`
}

// MoodSourceConfigFromViper marshals a MoodSourceConfig from v.
func MoodSourceConfigFromViper(v *viper.Viper) (*MoodSourceConfig, error) {
	if v = v.Sub("moodSource"); v == nil {
		v = viper.New()
	}
	var (
		cfg = new(MoodSourceConfig)
		err = v.Unmarshal(cfg, util.DecoderConfigOption)
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// SetDefaults zeroed fields in cfg to default values.
func (cfg *MoodSourceConfig) SetDefaults() {
	defaults.SetDefaults(cfg)
}

// A MoodSource implements job.MoodSource for an Airtable client.
type MoodSource struct {
	*airtable.Client
	cfg *MoodSourceConfig
}

func newMoodSource(c *airtable.Client, cfg *MoodSourceConfig) *MoodSource {
	return &MoodSource{
		Client: c,
		cfg:    cfg,
	}
}

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
		opts = airtable.FetchOptions{
			Limit: ms.cfg.Limit,
			Sort: []airtable.SortConfig{{
				Field:     "id",
				Direction: "desc",
			}},
		}
		records []*moodRecord
	)
	if err := ms.FetchRecords(MoodsTable, &records, &opts); err != nil {
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
