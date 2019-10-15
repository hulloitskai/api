package svcgql

import (
	"context"

	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/api/productivity"
	"go.stevenxie.me/api/productivity/prodgql"
	"go.stevenxie.me/gopkg/zero"
)

func newProductivityResolver() graphql.ProductivityResolver {
	return productivityResolver{}
}

type productivityResolver zero.Struct

var _ graphql.ProductivityResolver = (*productivityResolver)(nil)

func (productivityResolver) Score(
	_ context.Context,
	p *productivity.Productivity,
) (*int, error) {
	if p == nil {
		return nil, nil
	}
	intScore := int(*p.Score)
	return &intScore, nil
}

func newProductivityRecordResolver() graphql.ProductivityRecordResolver {
	return productivityRecordResolver{}
}

type productivityRecordResolver zero.Struct

var _ graphql.ProductivityRecordResolver = (*productivityRecordResolver)(nil)

func (res productivityRecordResolver) Category(
	_ context.Context,
	r *productivity.Record,
) (*prodgql.Category, error) {
	cat := r.Category
	return &prodgql.Category{
		ID:     int(cat),
		Name:   cat.Name(),
		Weight: int(cat.Weight()),
	}, nil
}

func (res productivityRecordResolver) Duration(
	_ context.Context,
	r *productivity.Record,
) (int, error) {
	return int(r.Duration.Seconds()), nil
}
