package gmaps

import (
	"net/http"
	"os"

	"github.com/cockroachdb/errors"

	"go.stevenxie.me/api/v2/pkg/google"
	"go.stevenxie.me/api/v2/pkg/httputil"
	"go.stevenxie.me/gopkg/name"
)

// NewTimelineClient creates a new TimelineClient.
//
// If it is not created with a custom http.Client, it will attempt to create
// one using the following environment variables:
//   - GOOGLE_HSID
//   - GOOGLE_SID
//   - GOOGLE_SSID
//
// These environment variables can be gleaned from the application cookies
// set by Google when you log in to Google Maps (check your web console for the
// cookies named 'SID', 'HSID', and 'SSID').
func NewTimelineClient(opts ...httputil.BasicClientOption) (TimelineClient, error) {
	var cfg httputil.BasicClientOptions
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.HTTPClient == nil { // derive authenticated client from envvars
		var (
			keys   = []string{"HSID", "SID", "SSID"}
			envmap = make(map[string]string, len(keys)) // key-val pairs
		)
		for _, k := range keys {
			var (
				key     = name.EnvKey(google.Namespace, k)
				val, ok = os.LookupEnv(key)
			)
			if !ok {
				return nil, errors.Newf("gmaps: no such environment variable '%s'", key)
			}
			envmap[k] = val
		}

		jar, err := cookiesFromMap(envmap)
		if err != nil {
			return nil, errors.Wrap(err, "gmaps: constructing cookies")
		}
		cfg.HTTPClient = &http.Client{Jar: jar}
	}

	return cfg.HTTPClient, nil
}

// A TimelineClient can make authenticated requests to the Google Maps
// timeline endpoint.
type TimelineClient httputil.BasicClient
