package gcal

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	gcal "google.golang.org/api/calendar/v3"

	"go.stevenxie.me/api/v2/pkg/timeutil"
	"go.stevenxie.me/api/v2/scheduling"
)

const _timeLayout = time.RFC3339

// NewCalendar creates a new scheduling.Calendar using a calendar.Service,
// which determines availability information using the Google calendars
// specified by ids.
func NewCalendar(
	svc *gcal.Service,
	ids []CalendarID,
) scheduling.Calendar {
	return calendar{
		svc: svc,
		ids: ids,
	}
}

// A CalendarID is a string.
type CalendarID = string

type calendar struct {
	svc *gcal.Service
	ids []CalendarID
}

var _ scheduling.Calendar = (*calendar)(nil)

func (cal calendar) BusyTimes(
	ctx context.Context,
	date time.Time,
) ([]scheduling.TimeSpan, error) {
	// Determine request time zone.
	tz := date.Location()

	// Build Busy request.
	var req gcal.FreeBusyRequest
	{
		req.Items = make([]*gcal.FreeBusyRequestItem, len(cal.ids))
		for i, id := range cal.ids {
			req.Items[i] = &gcal.FreeBusyRequestItem{Id: string(id)}
		}

		var (
			min = timeutil.DayStart(date)
			max = min.Add(time.Hour * 24)
		)
		req.TimeMin = min.Format(_timeLayout)
		req.TimeMax = max.Format(_timeLayout)
		req.TimeZone = tz.String()
	}

	// Perform request.
	res, err := cal.svc.Freebusy.Query(&req).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	// Parse availabilities.
	var periods []scheduling.TimeSpan
	for _, cal := range res.Calendars {
		// Catch errors in calendars response.
		if len(cal.Errors) > 0 {
			err = errors.New("gcal: error in calendar response")
			for _, cerr := range cal.Errors {
				err = errors.WithDetail(err, cerr.Reason)
			}
			return nil, err
		}

		for _, period := range cal.Busy {
			start, err := time.ParseInLocation(_timeLayout, period.Start, tz)
			if err != nil {
				return nil, errors.Wrap(err, "gcal: parsing start time")
			}
			end, err := time.ParseInLocation(_timeLayout, period.End, tz)
			if err != nil {
				return nil, errors.Wrap(err, "gcal: parsing end time")
			}
			periods = append(periods, scheduling.TimeSpan{
				Start: start,
				End:   end,
			})
		}
	}
	return periods, nil
}
