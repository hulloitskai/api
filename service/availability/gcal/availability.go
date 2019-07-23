package gcal

import (
	"sort"
	"time"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/service/availability"
	calendar "google.golang.org/api/calendar/v3"
)

// An AvailabilityService implements an availability.Service for a
// calendar.Service.
type AvailabilityService struct {
	calendar    *calendar.Service
	calendarIDs []string

	timezone *time.Location
}

var _ availability.Service = (*AvailabilityService)(nil)

// NewAvailabilityService creates a new AvailabilityService.
func NewAvailabilityService(
	calendar *calendar.Service,
	calendarIDs []string,
) *AvailabilityService {
	return &AvailabilityService{
		calendar:    calendar,
		calendarIDs: calendarIDs,
	}
}

// BusyPeriods determines periods of availability for a given date.
func (svc *AvailabilityService) BusyPeriods(date time.Time) (
	availability.Periods, error) {
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
	res, err := svc.calendar.Freebusy.Query(&req).Do()
	if err != nil {
		return nil, err
	}

	// Parse availabilities.
	busy := make(availability.Periods, 0)
	for _, cal := range res.Calendars {
		if len(cal.Errors) > 0 {
			err = errors.New("gcal: error in calendar response")
			for _, cerr := range cal.Errors {
				err = errors.WithDetail(err, cerr.Reason)
			}
		}
		for _, period := range cal.Busy {
			start, err := time.ParseInLocation(time.RFC3339, period.Start, timezone)
			if err != nil {
				return nil, errors.Wrap(err, "gcal: parsing start time")
			}
			end, err := time.ParseInLocation(time.RFC3339, period.End, timezone)
			if err != nil {
				return nil, errors.Wrap(err, "gcal: parsing end time")
			}
			busy = append(busy, &availability.Period{
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
		setting, err := svc.calendar.Settings.Get("timezone").Do()
		if err != nil {
			return nil, err
		}

		loc, err := time.LoadLocation(setting.Value)
		if err != nil {
			return nil, errors.Wrap(err, "gcal: failed to parse timezone")
		}
		svc.timezone = loc
	}

	return svc.timezone, nil
}

// sortablePeriods implements sort.Interface for availability.Periods.
type sortablePeriods availability.Periods

var _ sort.Interface = (*sortablePeriods)(nil)

func (sps sortablePeriods) Len() int      { return len(sps) }
func (sps sortablePeriods) Swap(i, j int) { sps[i], sps[j] = sps[j], sps[i] }
func (sps sortablePeriods) Less(i, j int) bool {
	if sps[i].Start.Before(sps[j].Start) {
		return true
	}
	if sps[i].Start.Equal(sps[j].Start) {
		return sps[i].End.Before(sps[j].End)
	}
	return false
}
