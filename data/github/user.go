package github

import (
	"encoding/json"

	errors "golang.org/x/xerrors"
)

// CurrentUserLogin gets the login of the authenticated user.
func (c *Client) CurrentUserLogin() (string, error) {
	res, err := c.httpc.Get(c.ghc.BaseURL.String() + "user")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var data struct {
		Login string `json:"login"`
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", errors.Errorf("github: decoding response as JSON: %w", err)
	}

	return data.Login, nil
}
