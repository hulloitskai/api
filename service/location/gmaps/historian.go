package gmaps

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stevenxie/api/pkg/google"
	"github.com/stevenxie/api/service/location"
)

type (
	// A Historian can load the authenticated user's location history from
	// Google Maps.
	Historian struct {
		client   *http.Client
		timezone *time.Location
	}

	// A HistorianConfig configures a Historian.
	HistorianConfig struct {
		Client   *http.Client
		Timezone *time.Location
	}
)

var _ location.HistoryService = (*Historian)(nil)

// NewHistorian creates a new Historian.
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
func NewHistorian(opts ...func(*HistorianConfig)) (*Historian, error) {
	var cfg HistorianConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Client == nil { // derive authenticated client from envvars
		var (
			vars   = []string{"HSID", "SID", "SSID"}
			varmap = make(map[string]string, len(vars))
		)
		for _, v := range vars {
			var (
				name    = fmt.Sprintf("%s_%s", strings.ToUpper(google.Namespace), v)
				val, ok = os.LookupEnv(name)
			)
			if !ok {
				return nil, errors.Newf("maps: no such envvar '%s'", name)
			}
			varmap[v] = val
		}

		jar, err := cookiesFromMap(varmap)
		if err != nil {
			return nil, errors.Wrap(err, "maps: constructing cookies")
		}
		cfg.Client = &http.Client{Jar: jar}
	}

	return &Historian{
		client:   cfg.Client,
		timezone: cfg.Timezone,
	}, nil
}
