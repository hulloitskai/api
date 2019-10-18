package svcgql

import (
	"context"
	"fmt"

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/gopkg/zero"
)

func newTransitDepartureResolver() graphql.TransitDepartureResolver {
	return transitDepartureResolver{}
}

type transitDepartureResolver zero.Struct

var _ graphql.TransitDepartureResolver = (*transitDepartureResolver)(nil)

func (res transitDepartureResolver) RelativeTimes(
	_ context.Context,
	d *transit.Departure,
) ([]string, error) {
	var (
		rts   = d.RelativeTimes()
		descs = make([]string, len(rts))
	)
	for i, rt := range rts {
		// Derive duration components.
		var (
			h = int(rt.Hours())
			m = int(rt.Minutes()) % 60
		)

		// Derive units.
		var (
			hourUnit = "hours"
			minUnit  = "minutes"
		)
		if h == 1 {
			hourUnit = "hour"
		}
		if m == 1 {
			minUnit = "minute"
		}

		if rt.Hours() > 1 {
			if m > 0 {
				descs[i] = fmt.Sprintf("%d %s and %d %s", h, hourUnit, m, minUnit)
			} else {
				descs[i] = fmt.Sprintf("%d %s", h, hourUnit)
			}
		} else {
			descs[i] = fmt.Sprintf("%d %s", m, minUnit)
		}
	}
	return descs, nil
}
