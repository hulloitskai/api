package maps

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	errors "golang.org/x/xerrors"

	"github.com/stevenxie/api/pkg/geo"
	"github.com/stevenxie/api/provider/google"
)

type (
	// A Historian can load the authenticated user's location history from
	// Google Maps.
	Historian struct {
		client *http.Client
	}

	// A HOption configures a Historian.
	HOption func(h *Historian)
)

var _ geo.LocationHistoryService = (*Historian)(nil)

// NewHistorian creates a new Historian.
//
// If it is not created with a custom http.Client, it will attempt to create
// one using the following envvars:
//   - GOOGLE_HSID
//   - GOOGLE_SID
//   - GOOGLE_SSID
//
// These environment variables can be gleaned from the application cookies
// set by Google when you log in to Google Maps (check your web console for the
// cookies named 'SID', 'HSID', and 'SSID').
func NewHistorian(opts ...HOption) (*Historian, error) {
	h := new(Historian)
	for _, opt := range opts {
		opt(h)
	}

	if h.client == nil { // derive authenticated client from envvars
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
				return nil, errors.Errorf("maps: no such envvar '%s'", name)
			}
			varmap[v] = val
		}

		jar, err := cookiesFromMap(varmap)
		if err != nil {
			return nil, errors.Errorf("maps: constructing cookies: %w", err)
		}
		h.client = &http.Client{Jar: jar}
	}

	return h, nil
}

// WithHClient configures a Historian to make HTTP requests with c.
func WithHClient(c *http.Client) HOption {
	return func(h *Historian) { h.client = c }
}
