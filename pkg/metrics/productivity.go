package metrics

// Productivity is a measurement of productivity, which describes the amount
// of time (in seconds) spent on distracting vs. productive activities.
type Productivity struct {
	VeryProductive  int `json:"veryProductive"`
	Productive      int `json:"productive"`
	Neutral         int `json:"neutral"`
	Distracting     int `json:"distracting"`
	VeryDistracting int `json:"veryDistracting"`
}

// A ProductivityService is able to retrieve the current productivity
// metrics.
type ProductivityService interface {
	CurrentProductivity() (*Productivity, error)
}
