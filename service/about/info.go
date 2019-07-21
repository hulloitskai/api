package about

import (
	"encoding/json"
	"fmt"
	"time"
)

// Info contains basic personal information.
type Info struct {
	Name        string        `json:"name"`
	Email       string        `json:"email"`
	Type        string        `json:"type"`
	Age         time.Duration `json:"age"`
	IQ          bool          `json:"iq"`
	Skills      []string      `json:"skills"`
	Whereabouts string        `json:"whereabouts"`
}

//revive:disable
func (i *Info) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name        string   `json:"name"`
		Email       string   `json:"email"`
		Type        string   `json:"type"`
		Age         string   `json:"age"`
		IQ          bool     `json:"iq"`
		Skills      []string `json:"skills"`
		Whereabouts string   `json:"whereabouts"`
	}{
		Name:        i.Name,
		Email:       i.Email,
		Type:        i.Type,
		Age:         fmt.Sprintf("about %d years", int(i.Age.Hours())/(365*24)),
		Skills:      i.Skills,
		Whereabouts: i.Whereabouts,
	})
}

// A Service can get personal information.
type Service interface {
	Info() (*Info, error)
}
