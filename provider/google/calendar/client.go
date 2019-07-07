package calendar

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	errors "golang.org/x/xerrors"

	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"gopkg.in/go-validator/validator.v2"

	"github.com/kelseyhightower/envconfig"
	googleprov "github.com/stevenxie/api/provider/google"
)

// A Client can interact with the Google Calendar API.
type Client struct{ calendar.Service }

// NewClient creates a new Client.
//
// It reads GOOGLE_TOKEN from the environment; if no such variable is found, an
// error will be returned.
func NewClient() (*Client, error) {
	var data struct {
		Token  string `validate:"nonzero"`
		ID     string `validate:"nonzero"`
		Secret string `validate:"nonzero"`
	}
	if err := envconfig.Process(googleprov.Namespace, &data); err != nil {
		return nil, errors.Errorf("calendar: reading envvars: %w", err)
	}

	// Validate data.
	if err := validator.Validate(&data); err != nil {
		return nil, errors.Errorf("calendar: validating envvars: %w", err)
	}

	// Create authenticated calendar service.
	var (
		config = oauth2.Config{
			ClientID:     data.ID,
			ClientSecret: data.Secret,
			Endpoint:     google.Endpoint,
		}
		token   = &oauth2.Token{RefreshToken: data.Token, TokenType: "Bearer"}
		tsource = config.TokenSource(context.Background(), token)
	)
	calsvc, err := calendar.NewService(
		context.Background(),
		option.WithTokenSource(tsource),
	)
	if err != nil {
		return nil, err
	}
	return &Client{*calsvc}, nil
}