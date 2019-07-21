package google

import (
	"context"

	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarService returns an authenticated calendar.Service.
func (cs *ClientSet) CalendarService(
	ctx context.Context,
	opts ...option.ClientOption,
) (*calendar.Service, error) {
	opts = append(opts, option.WithTokenSource(cs.source))
	return calendar.NewService(ctx, opts...)
}
