package github

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
)

// CurrentUserLogin gets the login of the authenticated user.
func (c *Client) CurrentUserLogin() (string, error) {
	if c.currentUserLogin == "" {
		res, err := c.httpc.Get(c.BaseURL() + "/user")
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		var data struct {
			Login string `json:"login"`
		}
		if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
			return "", errors.Wrap(err, "github: decoding response as JSON")
		}
		c.currentUserLogin = data.Login
	}
	return c.currentUserLogin, nil
}
