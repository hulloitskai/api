package prodgql

import (
	"context"

	"github.com/openlyinc/pointy"
	"go.stevenxie.me/api/v2/productivity"
	"go.stevenxie.me/gopkg/zero"
)

// A Resolver resolves fields for a productivity.Productivity.
type Resolver zero.Struct

//revive:disable-line:exported
func (Resolver) Score(
	_ context.Context,
	p *productivity.Productivity,
) (*int, error) {
	if p == nil {
		return nil, nil
	}
	return pointy.Int(int(*p.Score)), nil
}

// A RecordResolver resolves fields for a productivity.Record.
type RecordResolver zero.Struct

//revive:disable-line:exported
func (RecordResolver) Category(
	_ context.Context,
	r *productivity.Record,
) (*Category, error) {
	cat := r.Category
	return &Category{
		ID:     int(cat),
		Name:   cat.Name(),
		Weight: int(cat.Weight()),
	}, nil
}

//revive:disable
func (RecordResolver) Duration(_ context.Context, r *productivity.Record) (int, error) {
	return int(r.Duration.Seconds()), nil
}
