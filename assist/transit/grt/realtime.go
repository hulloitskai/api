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
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/assist/transit/transutil"
	"go.stevenxie.me/api/pkg/httputil"
)

const (
	_cacheMaxAge = 30 * time.Second
)

// NewRealtimeSource creates a transit.RealtimeSource that gets realtime data
// using GRT.
//
// If c == nil, a zero-value http.Client will be used.
func NewRealtimeSource(opts ...RealtimeSourceOption) (transit.RealtimeSource, error) {
	cfg := RealtimeSourceConfig{
		HTTPClient: new(http.Client),
		Logger:     logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Create log with component name.
	log := logutil.WithComponent(cfg.Logger, (*realtimeSource)(nil))

	// Use custom caching round-tripper.
	client := cfg.HTTPClient
	cache, err := httputil.NewCachingTripper(
		client.Transport,
		httputil.CachingTripperWithLogger(log),
		httputil.CachingTripperWithMaxAge(_cacheMaxAge),
	)
	if err != nil {
		return nil, errors.Wrap(err, "grt: creating CachingTripper")
	}
	client.Transport = cache

	return &realtimeSource{
		client: client,
		log:    log,
		tracer: cfg.Tracer,
		cache:  cache,
	}, nil
}

// WithLogger configures a transit.RealtimeSource to write logs with
// log.
func WithLogger(log *logrus.Entry) RealtimeSourceOption {
	return func(cfg *RealtimeSourceConfig) { cfg.Logger = log }
}

// WithTracer configures a transit.RealtimeSource to trace calls with t.
func WithTracer(t opentracing.Tracer) RealtimeSourceOption {
	return func(cfg *RealtimeSourceConfig) { cfg.Tracer = t }
}

type (
	realtimeSource struct {
		client *http.Client
		log    *logrus.Entry
		tracer opentracing.Tracer

		cache          *httputil.CachingTripper
		cacheTimestamp time.Time
	}

	// A RealtimeSourceConfig configures a transit.RealtimeSource.
	RealtimeSourceConfig struct {
		HTTPClient *http.Client
		Logger     *logrus.Entry
		Tracer     opentracing.Tracer
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
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, src.tracer,
		name.OfFunc((*realtimeSource).GetDepartureTimes),
	)
	defer span.Finish()

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
			Debug("Only GRT routes are supported.")
		return nil, transit.ErrOperatorNotSupported
	}

	// Get stop IDs.
	log.Trace("Getting corresponding stop IDs...")
	stopIDs, err := src.getStopIDs(ctx, &stn, &tp)
	if err != nil {
		log.
			WithError(errors.WithMessage(err, "grt")).
			Error("Failed to get corresponding stop IDs.")
		return nil, errors.Wrap(err, "grt: get corresponding stop IDs")
	}
	log = log.WithField("stop_ids", stopIDs)
	log.Trace("Got corresponding stop IDs.")

	// Get times for the correct stop ID.
	log.Trace("Getting departure times for stops...")
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
		log.Trace("No departure times found.")
		return []time.Time{}, nil
	}

	// ISSUE: Account for hiccups in the GRT realtime results, where incorrect
	// departure times (> 12 hours from current time) are returned.
	//
	// See: https://github.com/stevenxie/api/issues/3
	{
		var (
			now       = time.Now()
			anomalies = make([]time.Time, 0, len(times))
			filtered  = make([]time.Time, 0, len(times))
		)
		for _, t := range times {
			if t.Sub(now) > (12 * time.Hour) {
				anomalies = append(anomalies, t)
			} else {
				filtered = append(filtered, t)
			}
		}
		if len(anomalies) > 0 {
			log.
				WithField("anomalies", anomalies).
				Warn("Detected GRT departure time anomaly; removing affected results.")
			times = filtered
		}
	}

	// Sort times in ascending order.
	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	log.WithField("times", times).Trace("Sorted departure times.")
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
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, src.tracer,
		name.OfFunc((*realtimeSource).getStopIDs),
	)
	defer span.Finish()

	log := src.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod((*realtimeSource).getStopIDs),
		"station":         stn.Name,
		"route":           tp.Route,
		"direction":       tp.Direction,
	}).WithContext(ctx)

	// Build URL.
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
	url := u.String()

	// Perform request.
	log.WithField("url", url).Trace("Requesting stop data from GRT...")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		return nil, errors.Wrap(err, "decode response as JSON")
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "close response body")
	}
	log.
		WithField("response", stops).
		Trace("Decoded response data from GRT.")

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
	log.
		WithField("ids", ids).
		Trace("Got IDs.")

	return ids, nil
}

func (src *realtimeSource) getDepartureTimes(
	ctx context.Context,
	tp *transit.Transport,
	stopID string,
) ([]time.Time, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, src.tracer,
		name.OfFunc((*realtimeSource).getDepartureTimes),
	)
	defer span.Finish()

	log := src.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod((*realtimeSource).getDepartureTimes),
		"route":           tp.Route,
		"direction":       tp.Direction,
		"stop_id":         stopID,
	})

	// Build URL.
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

	url := u.String()
	log.WithField("url", url).Trace("Requesting stop info from GRT.")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		return nil, errors.Wrap(err, "decode JSON body")
	}
	log.WithField("response", data).Trace("Decoded response data.")

	// Marshal results to []time.Time.
	log.Trace("Marshalling response to []time.Time...")
	times := make([]time.Time, 0, len(data.StopTimes))
	for _, st := range data.StopTimes {
		log := log.WithFields(logrus.Fields{
			"arrival_date_time": st.ArrivalDateTime,
			"head_sign":         st.HeadSign,
		})

		// Ensure matching direction.
		dir := st.HeadSign
		if dir[1] == '-' {
			log.Trace("Head sign contains route component, parsing...")
			if c := dir[0]; c != tp.Route[len(tp.Route)-1] {
				log.Trace("Head sign route mismatch, skipping.")
				continue
			}
			dir = dir[2:]
		}
		if dir != tp.Direction {
			log.Trace("Wrong direction, skipping...")
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
		t := time.Unix(n/1000, n%1000)

		log.WithField("time", t).Trace("Parsed time from response date-time.")
		times = append(times, t)
	}
	return times, nil
}

func stripRouteDirection(route string) string {
	index := len(route) - 1
	if c := route[index]; ('A' <= c) && (c <= 'Z') {
		return route[:index]
	}
	return route
}
