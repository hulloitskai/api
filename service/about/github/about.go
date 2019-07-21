package github

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v25/github"
	"github.com/stevenxie/api/service/about"
	"github.com/stevenxie/api/service/location"
)

// AboutService reads an about.Info from a file stored in Github Gists.
type AboutService struct {
	gists            *github.GistsService
	gistID, gistFile string

	location location.Service
}

var _ about.Service = (*AboutService)(nil)

// NewAboutService creates a new AboutService that reads Info from a GitHub
// gist.
func NewAboutService(
	gists *github.GistsService,
	gistID, gistFile string,
	location location.Service,
) *AboutService {
	return &AboutService{
		gists:    gists,
		gistID:   gistID,
		gistFile: gistFile,
		location: location,
	}
}

// Info retrieves basic personal information from a GitHub gist.
func (svc *AboutService) Info() (*about.Info, error) {
	gist, _, err := svc.gists.Get(context.Background(), svc.gistID)
	if err != nil {
		return nil, err
	}

	var raw []byte
	for _, f := range gist.Files {
		if f.GetFilename() == svc.gistFile {
			if f.Content == nil {
				return nil, nil
			}
			raw = []byte(f.GetContent())
			break
		}
	}
	if raw == nil {
		return nil, errors.Newf("github: gist contains no such file '%s'",
			svc.gistFile)
	}

	// Decode file as JSON.
	var data struct {
		*about.Info
		Birthday string `json:"birthday"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		return nil, errors.Wrap(err, "github: decoding file contents as JSON")
	}

	// Derive age from birthday.
	bday, err := time.Parse("2006-01-02", data.Birthday)
	if err != nil {
		return nil, errors.Wrapf(err, "github: failed to parse birthday '%s'",
			data.Birthday)
	}
	data.Info.Age = time.Since(bday).Truncate(365 * 24 * time.Hour)

	// Fill in whereabouts using location service.
	data.Info.Whereabouts = "Unknown"
	if whereabouts, _ := svc.location.CurrentCity(); whereabouts != "" {
		data.Info.Whereabouts = whereabouts
	}
	return data.Info, nil
}
