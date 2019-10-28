package svcgql

import (
	"go.stevenxie.me/api/v2/graphql"
	"go.stevenxie.me/api/v2/location/locgql"
)

type locationResolvers struct {
	address        locgql.AddressResolver
	place          locgql.PlaceResolver
	historySegment locgql.HistorySegmentResolver
}

func (res locationResolvers) Place() graphql.PlaceResolver     { return res.place }
func (res locationResolvers) Address() graphql.AddressResolver { return res.address }
func (res locationResolvers) LocationHistorySegment() graphql.LocationHistorySegmentResolver {
	return res.historySegment
}
