package github

import (
	"encoding/json"
	"fmt"

	errors "golang.org/x/xerrors"
)

const (
	baseURL = "https://api.github.com"
	gistURL = baseURL + "/gists"
)

// GetGistFile reads a file stored on GitHub Gists.
func (c *Client) GetGistFile(id, file string) ([]byte, error) {
	res, err := c.httpc.Get(fmt.Sprintf("%s/%s", gistURL, id))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Decode response as JSON.
	var data struct {
		Files map[string]struct {
			Filename string `json:"filename"`
			Content  string `json:"content"`
		} `json:"files"`
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, errors.Errorf("decoding response as JSON: %w", err)
	}

	// Close response body.
	if err = res.Body.Close(); err != nil {
		return nil, errors.Errorf("closing response body: %w", err)
	}

	for _, f := range data.Files {
		if f.Filename == file {
			return []byte(f.Content), nil
		}
	}
	return nil, errors.Errorf("github: gist does not contain the file '%s'", file)
}
