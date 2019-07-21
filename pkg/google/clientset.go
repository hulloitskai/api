package google

import (
	"context"

	"golang.org/x/oauth2"
	googleauth "golang.org/x/oauth2/google"
	"gopkg.in/go-validator/validator.v2"

	"github.com/cockroachdb/errors"
	"github.com/kelseyhightower/envconfig"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "google"

// A ClientSet can create authenticated Google API clients.
type ClientSet struct{ source oauth2.TokenSource }

// NewClientSet creates a new ClientSet.
//
// It requires the following environment variables to be set:
//   - GOOGLE_ID
//   - GOOGLE_TOKEN
//   - GOOGLE_SECRET
func NewClientSet() (*ClientSet, error) {
	var data struct {
		Token  string `validate:"nonzero"`
		ID     string `validate:"nonzero"`
		Secret string `validate:"nonzero"`
	}
	if err := envconfig.Process(Namespace, &data); err != nil {
		return nil, errors.Wrap(err, "google: reading envvars")
	}

	// Validate data.
	if err := validator.Validate(&data); err != nil {
		return nil, errors.Wrap(err, "google: validating envvars")
	}

	// Create authenticated google service.
	var (
		config = oauth2.Config{
			ClientID:     data.ID,
			ClientSecret: data.Secret,
			Endpoint:     googleauth.Endpoint,
		}
		token  = &oauth2.Token{RefreshToken: data.Token, TokenType: "Bearer"}
		source = config.TokenSource(context.Background(), token)
	)
	return &ClientSet{source: source}, nil
}
