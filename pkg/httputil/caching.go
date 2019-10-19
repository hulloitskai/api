package httputil

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"
)

// NewCachingTripper creates a new CachingTripper that caches http.Responses
// from the underlying http.RoundTripper.
//
// If no underlying http.RoundTripper is provided, http.DefaultTransport will be
// used.
func NewCachingTripper(
	underlying http.RoundTripper,
	opts ...CachingTripperOption,
) *CachingTripper {
	if underlying == nil {
		underlying = http.DefaultTransport
	}
	cfg := CachingTripperConfig{
		Logger:      logutil.NoopEntry(),
		ExpiresFunc: nil,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &CachingTripper{
		Tripper:     underlying,
		expiresFunc: cfg.ExpiresFunc,
		cache:       make(map[string]cacheItem),
		log:         logutil.AddComponent(cfg.Logger, (*CachingTripper)(nil)),
	}
}

// WithCachingTripperLogger configures a CachingTripper to write logs with log.
func WithCachingTripperLogger(log *logrus.Entry) CachingTripperOption {
	return func(cfg *CachingTripperConfig) { cfg.Logger = log }
}

// WithCachingTripperMaxAge configures a CachingTripper to expire each response
// after age (the amount of time since the response was received).
func WithCachingTripperMaxAge(age time.Duration) CachingTripperOption {
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
		Tripper     http.RoundTripper
		expiresFunc ExpiresFunc
		log         *logrus.Entry

		mux   sync.Mutex
		cache map[string]cacheItem
	}

	// A CachingTripperConfig configures a CachingTripper.
	CachingTripperConfig struct {
		Logger      *logrus.Entry
		ExpiresFunc ExpiresFunc
	}

	// CachingTripperOption modifies a CachingTripperConfig.
	CachingTripperOption func(*CachingTripperConfig)

	// ExpiresFunc determines how the time at which a response for a request
	// expires.
	//
	// If the returned value is nil, the response will never expire. Similarly,
	// If ExpiresFunc is nil, responses never expire.
	ExpiresFunc func(req *http.Request, timestamp time.Time) (expires *time.Time)

	cacheItem struct {
		Response *http.Response
		Body     []byte
		Expires  *time.Time
	}
)

var _ http.RoundTripper = (*CachingTripper)(nil)

// Clear clears the CachingTransport's cache.
func (ct *CachingTripper) Clear() {
	ct.mux.Lock()
	ct.cache = make(map[string]cacheItem)
	ct.mux.Unlock()
}

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
	ct.mux.Lock()
	if item, ok := ct.cache[url]; ok {
		log := log.WithField("expires", item.Expires)
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

				ct.mux.Unlock()
				return res, nil
			}

			// Item is expired; remove it from cache.
			log.Trace("Cache item expired; removing...")
			delete(ct.cache, url)
		}
	}
	ct.mux.Unlock()

	// Cache miss, perform round-trip.
	log.Trace("Cache miss; performing round-trip")
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
	ct.mux.Lock()
	ct.cache[r.URL.String()] = item
	ct.mux.Unlock()
	return res, nil
}
