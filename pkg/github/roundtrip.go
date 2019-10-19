package github

import (
	"net/http"
	"net/textproto"

	funk "github.com/thoas/go-funk"
)

func newAcceptFilteringTripper(
	underlying http.RoundTripper,
	valuesToFilter ...string,
) http.RoundTripper {
	if underlying == nil {
		underlying = http.DefaultTransport
	}
	return acceptFilteringTripper{
		Tripper: underlying,
		Values:  valuesToFilter,
	}
}

type acceptFilteringTripper struct {
	Tripper http.RoundTripper
	Values  []string // headers to strip
}

var _ http.RoundTripper = (*acceptFilteringTripper)(nil)

var acceptMIMEHeaderKey = textproto.CanonicalMIMEHeaderKey("Accept")

func (aft acceptFilteringTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	accept := r.Header[acceptMIMEHeaderKey]
	if len(accept) > 0 {
		filtered := make([]string, 0, len(accept))
		for _, v := range accept {
			if !funk.ContainsString(aft.Values, v) {
				filtered = append(filtered, v)
			}
		}
		r.Header[acceptMIMEHeaderKey] = filtered
	}
	return aft.Tripper.RoundTrip(r)
}
