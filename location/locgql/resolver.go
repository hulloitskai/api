package locgql

import (
	"context"

	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/pkg/timeutil"
	"go.stevenxie.me/gopkg/zero"
)

// A PlaceResolver resolves fields for a location.Place.
type PlaceResolver zero.Struct

//revive:disable-line:exported
func (PlaceResolver) TimeZone(_ context.Context, p *location.Place) (*TimeZone, error) {
	tz := p.TimeZone
	if tz == nil {
		return nil, nil
	}
	name, offset := timeutil.CurrentZone(tz)
	return &TimeZone{
		Name:   name,
		Offset: offset,
	}, nil
}

// A HistorySegmentResolver resolves fields for a location.HistorySegment.
type HistorySegmentResolver zero.Struct

//revive:disable-line:exported
func (HistorySegmentResolver) Address(
	_ context.Context,
	seg *location.HistorySegment,
) (*string, error) {
	if seg.Address == "" {
		return nil, nil
	}
	return &seg.Address, nil
}

//revive:disable-line:exported
func (HistorySegmentResolver) Distance(
	_ context.Context,
	seg *location.HistorySegment,
) (*int, error) {
	if seg.Distance == 0 {
		return nil, nil
	}
	return &seg.Distance, nil
}

// An AddressResolver resolves fields for a location.Address.
type AddressResolver zero.Struct

//revive:disable-line:exported
func (res AddressResolver) District(
	_ context.Context,
	a *location.Address,
) (*string, error) {
	return res.resolveStringField(&a.District)
}

//revive:disable-line:exported
func (res AddressResolver) Street(
	_ context.Context,
	a *location.Address,
) (*string, error) {
	return res.resolveStringField(&a.Street)
}

//revive:disable-line:exported
func (res AddressResolver) Number(
	_ context.Context,
	a *location.Address,
) (*string, error) {
	return res.resolveStringField(&a.Number)
}

func (res AddressResolver) resolveStringField(v *string) (*string, error) {
	if *v == "" {
		return nil, nil
	}
	return v, nil
}
