package github

import (
	"context"

	"github.com/cockroachdb/errors"
)

// GistFile reads a file stored on GitHub Gists.
func (c *Client) GistFile(id, file string) ([]byte, error) {
	gist, _, err := c.ghc.Gists.Get(context.Background(), id)
	if err != nil {
		return nil, err
	}
	for _, f := range gist.Files {
		if f.GetFilename() == file {
			if f.Content == nil {
				return nil, nil
			}
			return []byte(f.GetContent()), nil
		}
	}
	return nil, errors.Newf("github: gist does not contain the file '%s'", file)
}
