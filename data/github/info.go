package github

import (
	"encoding/json"
	"time"

	"github.com/stevenxie/api/pkg/about"
	errors "golang.org/x/xerrors"
)

// InfoStore reads an about.Info from a file stored in Github Gists.
type InfoStore struct {
	repo         GistRepo
	gistID, file string
}

// A GistRepo can retrieve gist data.
type GistRepo interface {
	GistFile(id, file string) ([]byte, error)
}

var _ about.InfoService = (*InfoStore)(nil)

// NewInfoStore creates a new InfoStore that reads Info from a GitHub gist.
func NewInfoStore(gr GistRepo, gistID, file string) *InfoStore {
	return &InfoStore{
		repo:   gr,
		gistID: gistID,
		file:   file,
	}
}

// Info retrieves Info from a GitHub gist.
func (is *InfoStore) Info() (*about.Info, error) {
	raw, err := is.repo.GistFile(is.gistID, is.file)
	if err != nil {
		return nil, errors.Errorf("github: getting gist: %w", err)
	}

	// Decode gist contents.
	var data struct {
		*about.Info
		Birthday string `json:"birthday"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		return nil, errors.Errorf("github: decoding gist file contents as JSON: %w",
			err)
	}

	// Derive age from birthday.
	bday, err := time.Parse("2006-01-02", data.Birthday)
	if err != nil {
		return nil, errors.Errorf("github: failed to parse birthday '%s': %w",
			data.Birthday, err)
	}
	data.Info.Age = time.Since(bday).Truncate(365 * 24 * time.Hour)

	// Fill missing values.
	if data.Info.Whereabouts == "" {
		data.Info.Whereabouts = "Unknown (unimplemented)"
	}

	return data.Info, nil
}
