package airtable

// Config describes the options for configuring an airtable.Client.
type Config interface {
	APIKey() string
	BaseID() string
}
