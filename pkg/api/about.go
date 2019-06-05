package api

import (
	"encoding/json"
	"fmt"
	"time"
)

// About provides information about me.
type About struct {
	Name        string        `json:"name"`
	Email       string        `json:"email"`
	Type        string        `json:"type"`
	Age         time.Duration `json:"age"`
	IQ          bool          `json:"iq"`
	Skills      []string      `json:"skills"`
	Whereabouts string        `json:"whereabouts"`
}

// An AboutService can get About info.
type AboutService interface{ About() (*About, error) }

//revive:disable
func (a *About) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name        string   `json:"name"`
		Email       string   `json:"email"`
		Type        string   `json:"type"`
		Age         string   `json:"age"`
		IQ          bool     `json:"iq"`
		Skills      []string `json:"skills"`
		Whereabouts string   `json:"whereabouts"`
	}{
		Name:        a.Name,
		Email:       a.Email,
		Type:        a.Type,
		Age:         fmt.Sprintf("about %d years", int(a.Age.Hours())/(365*24)),
		Skills:      a.Skills,
		Whereabouts: a.Whereabouts,
	})
}
