package svcgql

import (
	"context"

	"go.stevenxie.me/api/about"
	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/gopkg/zero"
)

func newFullAboutResolver() graphql.FullAboutResolver {
	return fullAboutResolver{}
}

type fullAboutResolver zero.Struct

var _ graphql.FullAboutResolver = (*fullAboutResolver)(nil)

func (res fullAboutResolver) Age(
	_ context.Context,
	a *about.About,
) (string, error) {
	return a.Age.String(), nil
}
