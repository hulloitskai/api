package gcal

import (
	"context"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	errors "golang.org/x/xerrors"

	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/kelseyhightower/envconfig"
	"github.com/stevenxie/api/pkg/api"
	"gopkg.in/go-validator/validator.v2"
)

// A Client can interact with the Google Calendar API.
type Client struct {
	cs       *calendar.Service
	timezone *time.Location

	calendarIDs []string
}

var _ api.AvailabilityService = (*Client)(nil)

const envPrefix = "google"

// New creates a new Client.
//
// It reads GOOGLE_TOKEN from the environment; if no such variable is found, an
// error will be returned.
func New(calendarIDs []string) (*Client, error) {
	var data struct {
		Token  string `valid:"nonzero"`
		ID     string `valid:"nonzero"`
		Secret string `valid:"nonzero"`
	}
	if err := envconfig.Process(envPrefix, &data); err != nil {
		return nil, errors.Errorf("gcal: reading envvars: %w", err)
	}

	// Validate data.
	if err := validator.Validate(&data); err != nil {
		return nil, errors.Errorf("gcal: validating envvars: %w", err)
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
	return &Client{
		cs:          calsvc,
		calendarIDs: calendarIDs,
	}, nil
}
