package gqlutil

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	sentry "github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/zero"
)

// SentryRecoverFunc creates a new graphql.RecoverFunc that logs panics to
// Sentry.
func SentryRecoverFunc(
	hub *sentry.Hub,
	opts ...SentryRecoverOption,
) graphql.RecoverFunc {
	opt := SentryRecoverOptions{
		Logger:  logutil.NoopEntry(),
		Timeout: 2 * time.Second,
	}
	for _, apply := range opts {
		apply(&opt)
	}
	log := logutil.WithComponent(opt.Logger, SentryRecoverFunc)
	return func(ctx context.Context, err zero.Interface) error {
		id := hub.RecoverWithContext(ctx, err)
		if id != nil && opt.WaitForDelivery {
			hub.Flush(opt.Timeout)
		}
		log.WithField("id", id).Warn("Captured panic with Sentry.")
		return graphql.DefaultRecover(ctx, err)
	}
}

// SentryWithTimeout sets the timeout for event deliveries on a
// SentryRecoverFunc.
func SentryWithTimeout(t time.Duration) SentryRecoverOption {
	return func(opt *SentryRecoverOptions) { opt.Timeout = t }
}

// SentryWaitForDelivery configures a SentryRecoverFunc to wait for deliveries.
func SentryWaitForDelivery(wait bool) SentryRecoverOption {
	return func(opt *SentryRecoverOptions) { opt.WaitForDelivery = wait }
}

// SentryWithLogger configures a SentryRecoverFunc to writes logs with log.
func SentryWithLogger(log *logrus.Entry) SentryRecoverOption {
	return func(opt *SentryRecoverOptions) { opt.Logger = log }
}

type (
	// SentryRecoverOptions configures a SentryRecoverFunc.
	SentryRecoverOptions struct {
		Logger          *logrus.Entry
		WaitForDelivery bool
		Timeout         time.Duration
	}

	// A SentryRecoverOption modifies a SentryRecoverOptions.
	SentryRecoverOption func(*SentryRecoverOptions)
)
