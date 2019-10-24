package httputil

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dgraph-io/ristretto"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"
)

// NewCachingTripper creates a new CachingTripper that caches http.Responses
// from the underlying http.RoundTripper using an LRU cache.
//
// If no underlying http.RoundTripper is provided, http.DefaultTransport will be
// used.
func NewCachingTripper(
	underlying http.RoundTripper,
	opts ...CachingTripperOption,
) (*CachingTripper, error) {
	if underlying == nil {
		underlying = http.DefaultTransport
	}

	cfg := CachingTripperConfig{
		Logger: logutil.NoopEntry(),
		Ristretto: ristretto.Config{
			NumCounters: 64,             // keys to track frequency of
			MaxCost:     64 * (1 << 10), // max cache size, in bytes (64 KB)
			BufferItems: 64,
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	cache, err := ristretto.NewCache(&cfg.Ristretto)
	if err != nil {
		return nil, errors.Wrap(err, "httputil: creating Ristretto cache")
	}
	return &CachingTripper{
		Tripper: underlying,
		log:     logutil.WithComponent(cfg.Logger, (*CachingTripper)(nil)),

		expiresFunc: cfg.ExpiresFunc,
		cache:       cache,
	}, nil
}

// CachingTripperWithLogger configures a CachingTripper to write logs with log.
func CachingTripperWithLogger(log *logrus.Entry) CachingTripperOption {
	return func(cfg *CachingTripperConfig) { cfg.Logger = log }
}

// CachingTripperWithMaxAge configures a CachingTripper to expire each response
// after age (the amount of time since the response was received).
func CachingTripperWithMaxAge(age time.Duration) CachingTripperOption {
	return func(cfg *CachingTripperConfig) {
		cfg.ExpiresFunc = func(_ *http.Request, timestamp time.Time) *time.Time {
			t := timestamp.Add(age)
			return &t
		}
	}
}

type (
	// A CachingTripper is an http.RoundTripper that caches http.Responses based
	// on their request URL.
	CachingTripper struct {
		Tripper http.RoundTripper
		log     *logrus.Entry

		expiresFunc ExpiresFunc
		cache       *ristretto.Cache
	}

	// A CachingTripperConfig configures a CachingTripper.
	CachingTripperConfig struct {
		Logger      *logrus.Entry
		ExpiresFunc ExpiresFunc
		Ristretto   ristretto.Config
	}

	// CachingTripperOption modifies a CachingTripperConfig.
	CachingTripperOption func(*CachingTripperConfig)

	// ExpiresFunc determines the time at which the cached response for a request
	// must expire.
	//
	// If the returned value is nil, the response will not expire, and will be
	// cleared by the LRU cache once it reaches its maximum cost.
	//
	// Similarly, if ExpiresFunc is nil, no time-based expiry will be enforced
	// for any cached responeses.
	ExpiresFunc func(req *http.Request, timestamp time.Time) (expires *time.Time)

	cacheItem struct {
		Response *http.Response
		Body     []byte
		Expires  *time.Time
	}
)

var _ http.RoundTripper = (*CachingTripper)(nil)

// Clear clears the CachingTransport's cache.
func (ct *CachingTripper) Clear() { ct.cache.Clear() }

// RoundTrip implements http.RoundTripper for a CachingTripper.
func (ct *CachingTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	var (
		url = r.URL.String()
		log = ct.log.WithFields(logrus.Fields{
			logutil.MethodKey: name.OfMethod((*CachingTripper).RoundTrip),
			"url":             url,
		})
	)

	// Check cache.
	if v, ok := ct.cache.Get(url); ok {
		var (
			item = v.(cacheItem)
			log  = log.WithField("expires", item.Expires)
		)
		if e := item.Expires; e != nil {
			if e.After(time.Now()) {
				log := log
				if item.Body != nil {
					log = log.WithField("body", string(item.Body))
				}
				log.Trace("Cache hit; using cached response.")

				// Modify cached response body.
				res := item.Response
				res.Body = ioutil.NopCloser(bytes.NewReader(item.Body))

				return res, nil
			}

			// Item is expired; remove it from cache.
			log.Trace("Cache item expired; removing...")
			ct.cache.Del(url)
		}
	}

	// Cache miss, perform round-trip.
	log.Trace("Cache miss; performing round-trip.")
	res, err := ct.Tripper.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	timestamp := time.Now()

	// Copy response body.
	var (
		buf  = new(bytes.Buffer)
		body []byte
	)
	if b := res.Body; b != nil {
		if body, err = ioutil.ReadAll(io.TeeReader(res.Body, buf)); err != nil {
			log.WithError(err).Error("Reading response body.")
			return res, errors.Wrap(err, "httputil: read response body")
		}
		if err = b.Close(); err != nil {
			log.WithError(err).Error("Failed to close response body.")
			return res, errors.Wrap(err, "httputil: close response body")
		}
		res.Body = ioutil.NopCloser(buf)
	}
	{
		entry := log
		if body != nil {
			entry = entry.WithField("body", string(body))
		}
		entry.Trace("Got response from round-trip.")
	}

	// Create response item.
	item := cacheItem{
		Response: res,
		Body:     body,
	}
	if ct.expiresFunc != nil {
		item.Expires = ct.expiresFunc(r, timestamp)
	}
	log.
		WithField("expires", item.Expires).
		Trace("Created cache item.")

	// Cache response.
	ct.cache.Set(url, item, 1+int64(len(item.Body)))
	return res, nil
}
