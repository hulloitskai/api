package grt

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/assist/transit/transutil"
)

// NewRealtimeSource creates a transit.RealtimeSource that gets realtime data
// using GRT.
//
// If c == nil, http.DefaultClient will be used.
func NewRealtimeSource(opts ...RealtimeSourceOption) transit.RealtimeSource {
	cfg := RealtimeSourceConfig{
		HTTPClient: http.DefaultClient,
		Logger:     logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &realtimeSource{
		client: cfg.HTTPClient,
		log:    logutil.AddComponent(cfg.Logger, (*realtimeSource)(nil)),
	}
}

// WithRealtimeLogger configures a transit.RealtimeService to write logs with
// log.
func WithRealtimeLogger(log *logrus.Entry) RealtimeSourceOption {
	return func(cfg *RealtimeSourceConfig) { cfg.Logger = log }
}

type (
	realtimeSource struct {
		client *http.Client
		log    *logrus.Entry

		depResCache          map[string][]byte
		depResCacheTimestamp time.Time
	}

	// A RealtimeSourceConfig configures a transit.RealtimeService.
	RealtimeSourceConfig struct {
		Logger     *logrus.Entry
		HTTPClient *http.Client
	}

	// A RealtimeSourceOption modifies a RealtimeServiceConfig.
	RealtimeSourceOption func(*RealtimeSourceConfig)
)

var _ transit.RealtimeSource = (*realtimeSource)(nil)

// GetDepartureTime gets the realtime departure time for a given
// transit.Transport and transit.Station.
func (src *realtimeSource) GetDepartureTimes(
	ctx context.Context,
	tp transit.Transport,
	stn transit.Station,
) ([]time.Time, error) {
	log := logrus.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod((*realtimeSource).GetDepartureTimes),
		"route":           tp.Route,
		"direction":       tp.Direction,
		"station":         stn.Name,
	}).WithContext(ctx)

	// Check if operator is supported.
	if tp.Operator.Code != transit.OpCodeGRT {
		log.
			WithField("op_code", tp.Operator.Code).
			Warn("Only GRT routes are supported.")
		return nil, transit.ErrOperatorNotSupported
	}

	// Get stop IDs.
	stopIDs, err := src.getStopIDs(ctx, &stn, &tp)
	if err != nil {
		log.
			WithError(errors.WithMessage(err, "grt")).
			Error("Failed to get corresponding stop IDs.")
		return nil, errors.Wrap(err, "grt: get corresponding stop IDs")
	}
	log = log.WithField("stop_ids", stopIDs)
	log.Trace("Got stop IDs.")

	// Get times for the correct stop ID.
	var times []time.Time
	for _, id := range stopIDs {
		if times, err = src.getDepartureTimes(ctx, &tp, id); err != nil {
			err = errors.WithMessage(err, "grt")
			log.WithError(err).Error("Failed to get departure times.")
			return nil, err
		}
		if len(times) > 0 {
			break
		}
	}
	log = log.WithField("times", times)
	log.Trace("Got departure times.")

	// Break early if no times.
	if len(times) == 0 {
		return []time.Time{}, nil
	}

	// ISSUE: Account for hiccups in the GRT realtime results, where incorrect
	// departure times (> 12 hours from current time) are returned.
	//
	// See: https://github.com/stevenxie/api/issues/3
	{
		var (
			now     = time.Now()
			alerted bool
		)
		for i, t := range times {
			if t.Sub(now) > (12 * time.Hour) {
				if !alerted {
					log.Warn("Detected departure time anomaly from GRT. Removing " +
						"affected results.")
					alerted = true
				}

				// Remove this departure time.
				times = append(times[:i], times[i+1:]...)
			}
		}
	}

	// Sort times in ascending order.
	log.Trace("Sorting departure times...")
	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
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
	var (
		id = stripRouteDirection(tp.Route)
		qp = u.Query()
	)
	qp.Set("routeId", id)
	u.RawQuery = qp.Encode()

	// Perform request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}
	res, err := src.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Decode response.
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

	// Collect stop IDs.
	var ids []string
	for _, s := range stops {
		if transutil.NormalizeStationName(s.Name) == stn.Name {
			ids = append(ids, s.ID)
		}
	}
	if len(ids) == 0 {
		return nil, errors.WithDetailf(
			errors.New("none found"),
			"No GRT stops found with the name '%s'.", stn.Name,
		)
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
	var (
		id = stripRouteDirection(tp.Route)
		qp = u.Query()
	)
	qp.Set("routeId", id)
	qp.Set("stopId", stopID)
	u.RawQuery = qp.Encode()

	// If the cache hasn't been cleaned in 30 seconds, clean it!
	{
		now := time.Now()
		if src.depResCacheTimestamp.Before(now.Add(-30 * time.Second)) {
			src.depResCache = make(map[string][]byte)
			src.depResCacheTimestamp = now
		}
	}

	// Get stop times either from cache or from source.
	var (
		body []byte
		key  = stopID + tp.Route
		ok   bool
	)
	if body, ok = src.depResCache[key]; !ok {
		// Perform request.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, errors.Wrap(err, "create request")
		}
		res, err := src.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		// Read response body and cache it.
		// TODO: Use a generic HTTP caching client.
		if body, err = ioutil.ReadAll(res.Body); err != nil {
			return nil, errors.Wrap(err, "read response body")
		}
		if err = res.Body.Close(); err != nil {
			return nil, errors.Wrap(err, "close response body")
		}
		src.depResCache[key] = body
	}

	// Decode response as JSON.
	var data struct {
		StopTimes []struct {
			ArrivalDateTime string
			HeadSign        string
		}
	}
	if err = json.Unmarshal(body, &data); err != nil {
		return nil, errors.Wrap(err, "decoding JSON body")
	}

	// Marshal results to time.Time.
	times := make([]time.Time, 0, len(data.StopTimes))
	for _, st := range data.StopTimes {
		// Ensure matching direction.
		dir := st.HeadSign
		if dir[1] == '-' {
			if c := dir[0]; c != tp.Route[len(tp.Route)-1] {
				continue
			}
			dir = dir[2:]
		}
		if dir != tp.Direction {
			continue
		}

		// Parse date-time.
		s := strings.Trim(st.ArrivalDateTime, `\/`)
		if !strings.HasPrefix(s, "Date") {
			return nil, errors.Newf("unknown datetime format '%s'", s)
		}
		s = s[5 : len(s)-1]
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "converting datetime '%s' to int")
		}
		times = append(times, time.Unix(n/1000, n%1000))
	}
	return times, nil
}

func stripRouteDirection(route string) string {
	if c := route[len(route)-1]; ('A' <= c) && (c <= 'Z') {
		return route[:len(route)-1]
	}
	return route
}
