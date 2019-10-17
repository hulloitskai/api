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
			s = int(rt.Seconds()) % 60
		)

		// Derive units.
		var (
			hourUnit = "hour"
			minUnit  = "minute"
			secUnit  = "second"
		)
		if h > 1 {
			hourUnit = "hours"
		}
		if m > 1 {
			minUnit = "minutes"
		}
		if s > 1 {
			secUnit = "seconds"
		}

		const format = "%d %s and %d %s"
		if rt.Hours() > 1 {
			descs[i] = fmt.Sprintf(format, h, hourUnit, m, minUnit)
		} else {
			descs[i] = fmt.Sprintf(format, m, minUnit, s, secUnit)
		}
	}
	return descs, nil
}
