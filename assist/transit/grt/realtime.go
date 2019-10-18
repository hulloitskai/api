package grt

import (
	"context"
	"encoding/json"
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

// NewRealtimeService creates a transit.RealtimeSource that gets realtime data
// using GRT.
//
// If c == nil, http.DefaultClient will be used.
func NewRealtimeService(opts ...RealtimeServiceOption) transit.RealTimeService {
	cfg := RealtimeServiceConfig{
		HTTPClient: http.DefaultClient,
		Logger:     logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &realtimeService{
		client: cfg.HTTPClient,
		log:    logutil.AddComponent(cfg.Logger, (*realtimeService)(nil)),

		stopTimesCache: make(map[string][]string),
	}
}

// WithRealtimeLogger configures a transit.RealtimeService to write logs with
// log.
func WithRealtimeLogger(log *logrus.Entry) RealtimeServiceOption {
	return func(cfg *RealtimeServiceConfig) { cfg.Logger = log }
}

type (
	realtimeService struct {
		client *http.Client
		log    *logrus.Entry

		stopTimesCache          map[string][]string
		stopTimesCacheCleanTime time.Time
	}

	// A RealtimeServiceConfig configures a transit.RealtimeService.
	RealtimeServiceConfig struct {
		Logger     *logrus.Entry
		HTTPClient *http.Client
	}

	// A RealtimeServiceOption modifies a RealtimeServiceConfig.
	RealtimeServiceOption func(*RealtimeServiceConfig)
)

var _ transit.RealTimeService = (*realtimeService)(nil)

// GetDepartureTime gets the realtime departure time for a given
// transit.Transport and transit.Station.
func (src *realtimeService) GetDepartureTimes(
	ctx context.Context,
	tp transit.Transport,
	stn transit.Station,
) ([]time.Time, error) {
	log := logrus.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod((*realtimeService).GetDepartureTimes),
		"route":           tp.Route,
		"direction":       tp.Direction,
		"station":         stn.Name,
	})

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

func (src *realtimeService) getStopIDs(
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
		return nil, errors.Wrap(err, "create request")
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

func (src *realtimeService) getDepartureTimes(
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
			return nil, errors.Wrap(err, "create request")
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
			return nil, nil
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
