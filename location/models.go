package location

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/99designs/gqlgen/graphql"
	"go.stevenxie.me/gopkg/zero"
)

type (
	// A Place is a geographical location.
	Place struct {
		ID       string         `json:"id"`
		Level    string         `json:"level"`
		Type     string         `json:"type"`
		Position Coordinates    `json:"position"`
		TimeZone *time.Location `json:"timeZone,omitempty"`
		Address  Address        `json:"address"`
		Shape    []Coordinates  `json:"shape,omitempty"`
	}

	// An Address describes the position of a Place.
	Address struct {
		Label    string `json:"label"`
		Country  string `json:"country"`
		State    string `json:"state"`
		County   string `json:"county"`
		City     string `json:"city"`
		District string `json:"district,omitempty"`
		Postcode string `json:"postcode"`
		Street   string `json:"street,omitempty"`
		Number   string `json:"number,omitempty"`
	}
)

// Coordinates represents a point in 3D space.
//
// A GraphQL scalars, they are formatted as an array of the form [x,y]. They can
// be parsed in the forms [x,y], "[x,y]", or "x,y".
type Coordinates struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

var (
	_ graphql.Marshaler   = (*Coordinates)(nil)
	_ graphql.Unmarshaler = (*Coordinates)(nil)
)

//revive:disable-line:exported
func (c *Coordinates) UnmarshalGQL(v zero.Interface) error {
	// Parse base type.
	var parts []zero.Interface
	switch value := v.(type) {
	case []zero.Interface:
		parts = value
	case string:
		value = strings.Trim(value, "[]")
		fields := strings.Split(value, ",")
		for _, f := range fields {
			parts = append(parts, f)
		}
	default:
		return errors.New("location: coordinates must be either string or slice")
	}

	// Validate parts length.
	const nParts = 2
	if n := len(parts); n != nParts {
		return errors.Newf(
			"location: expected %d coordinate parts, but got %d",
			nParts, n,
		)
	}

	// Parse parts as float64s.
	components := make([]float64, len(parts))
	for i, part := range parts {
		if err := func() (err error) {
			switch v := part.(type) {
			case float64:
				components[i] = v
				return nil
			case string:
				v = strings.TrimSpace(v)
				components[i], err = strconv.ParseFloat(v, 64)
				return errors.WithMessage(err, "strconv")
			case json.Number:
				components[i], err = v.Float64()
				return errors.WithMessage(err, "float64")
			default:
				return errors.New("part is neither number nor string")
			}
		}(); err != nil {
			return errors.Wrapf(err, "location: parse coordinate part %d", i)
		}
	}

	c.X = components[0]
	c.Y = components[1]
	return nil
}

//revive:disable-line:exported
func (c Coordinates) MarshalGQL(w io.Writer) {
	fmt.Fprintf(w, `[%f,%f]`, c.X, c.Y)
}
