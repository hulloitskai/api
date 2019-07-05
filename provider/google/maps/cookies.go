package maps

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const (
	googleURL    = "https://google.com"
	cookieDomain = ".google.com"
	cookiePath   = "/"
)

// cookiesFromMap builds and fills a cookiejar.Jar with the name-value pairs
// of cookie data in m.
func cookiesFromMap(m map[string]string) (*cookiejar.Jar, error) {
	url, err := url.Parse(googleURL)
	if err != nil {
		panic(err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	cookies := make([]*http.Cookie, 0, len(m))
	for k, v := range m {
		cookies = append(cookies, &http.Cookie{
			Name:   k,
			Value:  v,
			Domain: cookieDomain,
			Path:   cookiePath,
			Secure: true,
		})
	}
	jar.SetCookies(url, cookies)

	return jar, nil
}
