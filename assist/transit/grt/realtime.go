package grt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/assist/transit"
)

// NewRealtimeSource creates a transit.RealtimeSource that gets realtime data
// using GRT.
//
// If c == nil, http.DefaultClient will be used.
func NewRealtimeSource(c *http.Client) transit.RealtimeSource {
	if c == nil {
		c = http.DefaultClient
	}
	return &realtimeSource{
		client:         c,
		stopTimesCache: make(map[string][]string),
	}
}

type realtimeSource struct {
	client *http.Client

	stopTimesCache          map[string][]string
	stopTimesCacheCleanTime time.Time
}

var _ transit.RealtimeSource = (*realtimeSource)(nil)

// GetDepartureTime gets the realtime departure time for a given
// transit.Transport and transit.Station.
func (src *realtimeSource) GetDepartureTimes(
	ctx context.Context,
	tp transit.Transport,
	stn transit.Station,
) ([]time.Time, error) {
	// Check if operator is supported.
	if tp.Operator.Code != transit.OpCodeGRT {
		return nil, transit.ErrOperatorNotSupported
	}
	stopIDs, err := src.getStopIDs(ctx, &stn, &tp)
	if err != nil {
		return nil, errors.Wrap(err, "grt: get matching stop IDs")
	}
	var times []time.Time
	for _, id := range stopIDs {
		if times, err = src.getDepartureTimes(ctx, &tp, id); err != nil {
			return nil, errors.WithMessage(err, "grt")
		}
		if len(times) > 0 {
			break
		}
	}
	return times, nil
}

const (
	_baseURL = "https://realtimemap.grt.ca"
	_stopURL = _baseURL + "/Stop/GetByRouteId"
	_depsURL = _baseURL + "/Stop/GetStopInfo"
)

func (src *realtimeSource) getStopIDs(
	ctx context.Context,
	stn *transit.Station,
	tp *transit.Transport,
) ([]string, error) {
	// Construct URL.
	u, err := url.Parse(_stopURL)
	if err != nil {
		panic(err)
	}
	qp := u.Query()
	qp.Set("routeId", tp.Route)
	u.RawQuery = qp.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}
	res, err := src.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var stops []struct {
		ID   string `json:"StopId"`
		Name string
	}
	if err = json.NewDecoder(res.Body).Decode(&stops); err != nil {
		return nil, errors.Wrap(err, "decoding response as JSON")
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "closing response body")
	}

	var ids []string
	for _, s := range stops {
		if s.Name == stn.Name {
			ids = append(ids, s.ID)
		}
	}
	if len(ids) == 0 {
		return nil, errors.New("none found")
	}
	return ids, nil
}

func (src *realtimeSource) getDepartureTimes(
	ctx context.Context,
	tp *transit.Transport,
	stopID string,
) ([]time.Time, error) {
	// Construct URL.
	u, err := url.Parse(_depsURL)
	if err != nil {
		panic(err)
	}
	qp := u.Query()
	qp.Set("routeId", tp.Route)
	qp.Set("stopId", stopID)
	u.RawQuery = qp.Encode()

	// If the cache hasn't been cleaned in 30 seconds, clean it!
	{
		now := time.Now()
		if src.stopTimesCacheCleanTime.Before(now.Add(-30 * time.Second)) {
			src.stopTimesCache = make(map[string][]string)
			src.stopTimesCacheCleanTime = now
		}
	}

	// Get stop times either from cache or from source.
	var (
		stopTimes []string
		key       = stopID + tp.Route
		ok        bool
	)
	if stopTimes, ok = src.stopTimesCache[key]; !ok {
		// Perform request.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, errors.Wrap(err, "creating request")
		}
		res, err := src.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		// Decode response as JSON.
		var data struct {
			StopTimes []struct {
				ArrivalDateTime string
				HeadSign        string
			}
		}
		if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
			return nil, errors.Wrap(err, "decoding response as JSON")
		}
		if err = res.Body.Close(); err != nil {
			return nil, errors.Wrap(err, "closing response body")
		}

		stopTimes = make([]string, len(data.StopTimes))
		for i, st := range data.StopTimes {
			stopTimes[i] = st.ArrivalDateTime
		}

		src.stopTimesCache[key] = stopTimes

		// Return early if no results or wrong direction.
		if len(stopTimes) == 0 {
			return nil, errors.New("no results")
		}
		if data.StopTimes[0].HeadSign != tp.Direction {
			return []time.Time{}, nil
		}
	}

	deps := make([]time.Time, len(stopTimes))
	for i, st := range stopTimes {
		s := strings.Trim(st, `\/`)
		if !strings.HasPrefix(s, "Date") {
			return nil, errors.Newf("unknown datetime format '%s'", s)
		}
		s = s[5 : len(s)-1]
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "converting datetime '%s' to int")
		}
		deps[i] = time.Unix(n/1000, n%1000)
	}
	return deps, nil
}
