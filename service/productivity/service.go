package productivity

type (
	// Period is a period of time associated with a degree of productivity /
	// unproductivity.
	Period struct {
		Name     string `json:"name"`
		ID       int    `json:"id"`
		Duration int    `json:"duration"`
	}

	// Periods are a set of Periods.
	Periods []*Period
)

// A Service is able to retrieve the current productivity metrics.
type Service interface {
	CurrentProductivity() (Periods, error)
}
