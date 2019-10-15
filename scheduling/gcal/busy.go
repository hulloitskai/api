package gcal

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	calendar "google.golang.org/api/calendar/v3"

	"go.stevenxie.me/api/pkg/timeutil"
	"go.stevenxie.me/api/scheduling"
)

const _timeLayout = time.RFC3339

// NewBusySource creates a new scheduling.BusySource using a calendar.Service,
// which determines availability information using the calendars specified by
// calIDs.
func NewBusySource(
	calsvc *calendar.Service,
	calIDs []string,
) scheduling.BusySource {
	return busySource{
		calsvc: calsvc,
		cids:   calIDs,
	}
}

type busySource struct {
	calsvc *calendar.Service
	cids   []string
}

var _ scheduling.BusySource = (*busySource)(nil)

func (svc busySource) BusyPeriods(
	ctx context.Context,
	date time.Time,
) ([]scheduling.TimePeriod, error) {
	// Determine request time zone.
	tz := date.Location()

	// Build Busy request.
	var req calendar.FreeBusyRequest
	{
		req.Items = make([]*calendar.FreeBusyRequestItem, len(svc.cids))
		for i, id := range svc.cids {
			req.Items[i] = &calendar.FreeBusyRequestItem{Id: id}
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
	res, err := svc.calsvc.Freebusy.Query(&req).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	// Parse availabilities.
	var periods []scheduling.TimePeriod
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
			periods = append(periods, scheduling.TimePeriod{
				Start: start,
				End:   end,
			})
		}
	}
	return periods, nil
}
