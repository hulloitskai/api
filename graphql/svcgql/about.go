package svcgql

import (
	"context"
	"fmt"

	"go.stevenxie.me/api/about"
	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/gopkg/zero"
)

func newFullAboutResolver() graphql.FullAboutResolver {
	return fullAboutResolver{}
}

type fullAboutResolver zero.Struct

var _ graphql.FullAboutResolver = (*fullAboutResolver)(nil)

func (res fullAboutResolver) Birthday(
	_ context.Context,
	a *about.About,
) (string, error) {
	return a.Birthday.Format("2006-01-02"), nil
}

func (res fullAboutResolver) Age(
	_ context.Context,
	a *about.About,
) (string, error) {
	return fmt.Sprintf("%d days", int(a.Age.Hours())/24), nil
}
