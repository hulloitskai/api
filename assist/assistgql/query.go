package assistgql

import (
	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/assist/transit/transgql"
)

// NewQuery creates a new Query.
func NewQuery(svcs QueryServices) Query {
	return Query{
		Transit: transgql.NewQuery(svcs.Transit),
	}
}

type (
	// A Query resolves queries for transit-related data.
	Query struct {
		Transit transgql.Query `json:"transit"`
	}

	// QueryServices are services used by Query to resolve queries.
	QueryServices struct {
		Transit transit.Service
	}
)
