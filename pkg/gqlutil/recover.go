package gqlutil

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	sentry "github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/zero"
)

// SentryRecoverFunc creates a new graphql.RecoverFunc that logs panics to
// Sentry.
func SentryRecoverFunc(c *sentry.Client, log *logrus.Entry) graphql.RecoverFunc {
	log = logutil.WithComponent(log, SentryRecoverFunc)
	scope := sentry.NewScope()
	return func(ctx context.Context, err zero.Interface) error {
		id := c.RecoverWithContext(
			ctx,
			err, &sentry.EventHint{RecoveredException: err},
			scope,
		)
		log.WithField("id", id).Warn("Captured panic with Sentry.")
		return graphql.DefaultRecover(ctx, err)
	}
}
