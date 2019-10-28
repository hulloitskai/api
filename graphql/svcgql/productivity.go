package svcgql

import (
	"go.stevenxie.me/api/v2/graphql"
	"go.stevenxie.me/api/v2/productivity/prodgql"
)

type productivityResolvers struct {
	productivity prodgql.Resolver
	record       prodgql.RecordResolver
}

func (res productivityResolvers) Productivity() graphql.ProductivityResolver {
	return res.productivity
}

func (res productivityResolvers) ProductivityRecord() graphql.ProductivityRecordResolver {
	return res.record
}
