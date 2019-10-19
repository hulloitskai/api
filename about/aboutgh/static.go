package aboutgh

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v25/github"
	"go.stevenxie.me/api/about"
)

// NewStaticSource creates a new about.StaticSource that reads from a GitHub
// gist.
func NewStaticSource(
	gists *github.GistsService,
	gistID, gistFile string,
) about.StaticSource {
	return staticSource{
		gists:    gists,
		gistID:   gistID,
		gistFile: gistFile,
	}
}

type staticSource struct {
	gists            *github.GistsService
	gistID, gistFile string
}

var _ about.StaticSource = (*staticSource)(nil)

func (src staticSource) GetStatic() (*about.Static, error) {
	gist, _, err := src.gists.Get(context.Background(), src.gistID)
	if err != nil {
		return nil, err
	}

	var raw []byte
	for _, f := range gist.Files {
		if f.GetFilename() == src.gistFile {
			if f.Content == nil {
				return nil, nil
			}
			raw = []byte(f.GetContent())
			break
		}
	}
	if raw == nil {
		return nil, errors.Newf("aboutgh: gist contains no such file '%s'",
			src.gistFile)
	}

	// Decode file as JSON.
	var data struct {
		*about.Static
		Birthday string `json:"birthday"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		return nil, errors.Wrap(err, "aboutgh: decode file contents as JSON")
	}

	// Derive age from birthday.
	data.Static.Birthday, err = time.Parse("2006-01-02", data.Birthday)
	if err != nil {
		return nil, errors.Wrapf(err, "aboutgh: failed to parse birthday '%s'",
			data.Birthday)
	}
	return data.Static, nil
}
