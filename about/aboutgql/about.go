package aboutgql

import (
	"context"
	"fmt"

	"go.stevenxie.me/api/about"
	"go.stevenxie.me/gopkg/zero"
)

// A Resolver resolves fields for an about.About.
type Resolver zero.Struct

//revive:disable-line:exported
func (Resolver) Birthday(_ context.Context, a *about.About) (string, error) {
	return a.Birthday.Format("2006-01-02"), nil
}

//revive:disable-line:exported
func (Resolver) Age(_ context.Context, a *about.About) (string, error) {
	return fmt.Sprintf("%d days", int(a.Age.Hours())/24), nil
}
