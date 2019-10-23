package locgql

// A TimeZone represents a time zone.
type TimeZone struct {
	Name string `json:"name"`

	// The offset in seconds east of UTC.
	Offset int `json:"offset"`
}
