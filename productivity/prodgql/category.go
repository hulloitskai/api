package prodgql

// A Category is a GraphQL representation of a productivity.Category.
type Category struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}
