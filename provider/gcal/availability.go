package gcal

import (
	"sort"
	"time"

	errors "golang.org/x/xerrors"
	calendar "google.golang.org/api/calendar/v3"

	"github.com/stevenxie/api/pkg/api"
)

// BusyPeriods determines periods of availability for a given date.
func (c *Client) BusyPeriods(date time.Time) ([]*api.TimePeriod, error) {
	timezone, err := c.Timezone()
	if err != nil {
		return nil, err
	}

	// Determine calendars to query.
	items := make([]*calendar.FreeBusyRequestItem, len(c.calendarIDs))
	for i, id := range c.calendarIDs {
		items[i] = &calendar.FreeBusyRequestItem{Id: id}
	}

	// Build FreeBusy request.
	var (
		min = time.Date(
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
	res, err := c.cs.Freebusy.Query(&req).Do()
	if err != nil {
		return nil, err
	}

	// Parse availabilities.
	busy := make([]*api.TimePeriod, 0)
	for _, cal := range res.Calendars {
		if len(cal.Errors) > 0 {
			return nil, errors.Errorf("gcal: error in calendars response: %w",
				cal.Errors[0])
		}
		for _, period := range cal.Busy {
			start, err := time.ParseInLocation(time.RFC3339, period.Start, timezone)
			if err != nil {
				return nil, errors.Errorf("gcal: parsing start time: %w", err)
			}
			end, err := time.ParseInLocation(time.RFC3339, period.End, timezone)
			if err != nil {
				return nil, errors.Errorf("gcal: parsing end time: %w", err)
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
func (c *Client) Timezone() (*time.Location, error) {
	if c.timezone == nil {
		setting, err := c.cs.Settings.Get("timezone").Do()
		if err != nil {
			return nil, err
		}

		loc, err := time.LoadLocation(setting.Value)
		if err != nil {
			return nil, errors.Errorf("gcal: failed to parse timezone: %w", err)
		}
		c.timezone = loc
	}

	return c.timezone, nil
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
