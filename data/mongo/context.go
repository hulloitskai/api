package mongo

import "context"

type contextGenerator func() (context.Context, context.CancelFunc)
