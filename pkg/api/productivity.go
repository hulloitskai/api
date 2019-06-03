package api

// ProductivitySegment is a measurement of productivity in terms of time spent
// doing something productive / unproductive.
type ProductivitySegment struct {
	Name     string `json:"name"`
	ID       int    `json:"id"`
	Duration int    `json:"duration"`
}

// A ProductivityService is able to retrieve the current productivity metrics.
type ProductivityService interface {
	CurrentProductivity() ([]*ProductivitySegment, error)
}
