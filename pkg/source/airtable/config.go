package airtable

// Config describes the options for configuring an airtable.Client.
type Config struct {
	APIKey string `env:"API_KEY"`
	BaseID string `env:"BASE_ID"`

	MoodTableName string `env:"MOOD_TABLE"`
	MoodTableView string `env:"MOOD_VIEW"`
}

func (cfg *Config) configureDefaults() {
	if cfg.MoodTableName == "" {
		cfg.MoodTableName = "moods"
	}
}
