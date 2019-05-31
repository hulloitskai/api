package about

import (
	"encoding/json"
	"fmt"
	"time"
)

// Info provides information about me.
type Info struct {
	Name        string
	Type        string
	Age         time.Duration
	IQ          bool
	Skills      []string
	Whereabouts string
}

// An InfoService can get Info.
type InfoService interface {
	Info() (*Info, error)
}

//revive:disable
func (i *Info) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name        string   `json:"name"`
		Type        string   `json:"type"`
		Age         string   `json:"age"`
		IQ          bool     `json:"iq"`
		Skills      []string `json:"skills"`
		Whereabouts string   `json:"whereabouts"`
	}{
		Name:        i.Name,
		Type:        i.Type,
		Age:         fmt.Sprintf("about %d years", int(i.Age.Hours())/(365*24)),
		Skills:      i.Skills,
		Whereabouts: i.Whereabouts,
	})
}
