package about

import "time"

// Static is personal information that does not vary based on external
// factors.
type Static struct {
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Type     string    `json:"type"`
	Birthday time.Time `json:"birthday"`
	IQ       bool      `json:"iq"`
	Skills   []string  `json:"skills"`
}

// A StaticSource is a source of StaticInfos.
type StaticSource interface {
	GetStatic() (*Static, error)
}
