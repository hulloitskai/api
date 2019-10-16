package about

import (
	"time"

	"go.stevenxie.me/api/location"
)

type (
	// About contains basic personal information.
	About struct {
		Name     string               `json:"name"`
		Email    string               `json:"email"`
		Type     string               `json:"type"`
		Birthday time.Time            `json:"birthday"`
		Age      time.Duration        `json:"age"`
		IQ       bool                 `json:"iq"`
		Skills   []string             `json:"skills"`
		Location location.Coordinates `json:"location"`
	}

	// Masked contains obfuscated basic personal information.
	Masked struct {
		Name        string
		Email       string
		Type        string
		ApproxAge   string
		IQ          bool
		Skills      []string
		Whereabouts string
	}
)

var (
	_ ContactInfo = (*About)(nil)
	_ ContactInfo = (*Masked)(nil)
)

// GetName implements ContactInfo.GetName for an About.
func (a *About) GetName() string { return a.Name }

// GetEmail implements ContactInfo.GetEmail for an About.
func (a *About) GetEmail() string { return a.Email }

// GetName implements ContactInfo.GetName for a Masked.
func (m *Masked) GetName() string { return m.Name }

// GetEmail implements ContactInfo.GetEmail for a Masked.
func (m *Masked) GetEmail() string { return m.Email }

// ContactInfo contains my name and email.
type ContactInfo interface {
	GetName() string
	GetEmail() string
}
