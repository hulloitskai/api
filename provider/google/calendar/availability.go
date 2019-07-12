package calendar

import (
	"sort"
	"time"

	calendar "google.golang.org/api/calendar/v3"

	"github.com/cockroachdb/errors"
	"github.com/stevenxie/api/pkg/api"
)

// An AvailabilityService implements an api.AvailabilityService for a Client.
type AvailabilityService struct {
	client      *Client
	calendarIDs []string

	timezone *time.Location
}

var _ api.AvailabilityService = (*AvailabilityService)(nil)

// NewAvailabilityService creates a new AvailabilityService.
func NewAvailabilityService(
	c *Client,
	calendarIDs []string,
) *AvailabilityService {
	return &AvailabilityService{
		client:      c,
		calendarIDs: calendarIDs,
	}
}

// BusyPeriods determines periods of availability for a given date.
func (svc *AvailabilityService) BusyPeriods(date time.Time) ([]*api.TimePeriod,
	error) {
	// Determine calendars to query.
	items := make([]*calendar.FreeBusyRequestItem, len(svc.calendarIDs))
	for i, id := range svc.calendarIDs {
		items[i] = &calendar.FreeBusyRequestItem{Id: id}
	}

	// Build FreeBusy request.
	var (
		timezone = date.Location()
		min      = time.Date(
			date.Year(), date.Month(), date.Day(),
			0, 0, 0, 0,
			date.Location(),
		)
		req = calendar.FreeBusyRequest{
			Items:    items,
			TimeMin:  min.Format(time.RFC3339),
			TimeMax:  min.Add(time.Hour * 24).Format(time.RFC3339),
			TimeZone: timezone.String(),
		}
	)

	// Perform request.
	res, err := svc.client.Freebusy.Query(&req).Do()
	if err != nil {
		return nil, err
	}

	// Parse availabilities.
	busy := make([]*api.TimePeriod, 0)
	for _, cal := range res.Calendars {
		if len(cal.Errors) > 0 {
			err = errors.New("calendar: error in calendar response")
			for _, cerr := range cal.Errors {
				err = errors.WithDetail(err, cerr.Reason)
			}
		}
		for _, period := range cal.Busy {
			start, err := time.ParseInLocation(time.RFC3339, period.Start, timezone)
			if err != nil {
				return nil, errors.Wrap(err, "calendar: parsing start time")
			}
			end, err := time.ParseInLocation(time.RFC3339, period.End, timezone)
			if err != nil {
				return nil, errors.Wrap(err, "calendar: parsing end time")
			}
			busy = append(busy, &api.TimePeriod{
				Start: start,
				End:   end,
			})
		}
	}

	sort.Sort(sortablePeriods(busy))
	return busy, nil
}

// Timezone returns the authenticated user's timezone.
func (svc *AvailabilityService) Timezone() (*time.Location, error) {
	if svc.timezone == nil {
		setting, err := svc.client.Settings.Get("timezone").Do()
		if err != nil {
			return nil, err
		}

		loc, err := time.LoadLocation(setting.Value)
		if err != nil {
			return nil, errors.Wrap(err, "calendar: failed to parse timezone")
		}
		svc.timezone = loc
	}

	return svc.timezone, nil
}

// sortablePeriods implements sort.Interface for a slice of
// api.TimePeriods.
type sortablePeriods []*api.TimePeriod

var _ sort.Interface = (*sortablePeriods)(nil)

func (sp sortablePeriods) Len() int      { return len(sp) }
func (sp sortablePeriods) Swap(i, j int) { sp[i], sp[j] = sp[j], sp[i] }
func (sp sortablePeriods) Less(i, j int) bool {
	if sp[i].Start.Before(sp[j].Start) {
		return true
	}
	if sp[i].Start.Equal(sp[j].Start) {
		return sp[i].End.Before(sp[j].End)
	}
	return false
}
