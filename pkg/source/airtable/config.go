package airtable

// Config describes the options for configuring an airtable.Client.
type Config struct {
	APIKey string `env:"API_KEY"`
	BaseID string `env:"BASE_ID"`

	MoodTableName string `env:"MOOD_TABLE"`
	MoodTableView string `env:"MOOD_VIEW"`
}

func (cfg *Config) setDefaults() {
	if cfg.MoodTableName == "" {
		cfg.MoodTableName = "moods"
	}
	if cfg.MoodTableView == "" {
		cfg.MoodTableView = "Grid view"
	}
}
