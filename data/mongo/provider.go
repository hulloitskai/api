package mongo

import (
	"context"
	"time"

	errors "golang.org/x/xerrors"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/stevenxie/api/internal/util"
)

// A Provider provides various services that use Mongo as an underlying data
// store.
type Provider struct {
	client *mongo.Client
	db     *mongo.Database
	dbname string

	moodService *MoodService

	connectTimeout   time.Duration
	operationTimeout time.Duration
}

// NewProvider creates a new Provider that connects to Mongo using uri, and
// uses the specified db.
//
// uri should be of the format:
//   mongodb://<username>:<password>@<host>:<port>[/<database>]
func NewProvider(uri, db string) (*Provider, error) {
	// Validate arguments.
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	if db == "" {
		return nil, errors.New("mongo: empty database name")
	}

	// Create Mongo client.
	client, err := mongo.NewClient(uri)
	if err != nil {
		return nil, err
	}

	return &Provider{
		client: client,
		dbname: db,
	}, nil
}

// SetConnectTimeout sets the connect / disconnect timeout.
func (p *Provider) SetConnectTimeout(timeout time.Duration) {
	p.connectTimeout = timeout
}

// SetOperationTimeout sets the default timeout for a Mongo operation.
func (p *Provider) SetOperationTimeout(timeout time.Duration) {
	p.operationTimeout = timeout
	if p.moodService != nil {
		p.moodService.SetTimeout(timeout)
	}
}

func (p *Provider) connectionContext() (context.Context, context.CancelFunc) {
	return util.ContextWithTimeout(p.connectTimeout)
}

// Open opens a connection to Mongo, and initializes internal services.
func (p *Provider) Open() error {
	// Connect to DB.
	ctx, cancel := p.connectionContext()
	defer cancel()
	if err := p.client.Connect(ctx); err != nil {
		return err
	}

	// Initialize p.db.
	db := p.client.Database(p.dbname)
	if db == nil {
		return errors.Errorf("mongo: no such database '%s'", p.dbname)
	}
	p.db = db

	// Initialize services.
	moods := newMoodService(db)
	moods.SetTimeout(p.operationTimeout)
	if err := moods.init(); err != nil {
		return errors.Errorf("mongo: initializing mood service: %w", err)
	}
	p.moodService = moods
	return nil
}

// Close closes the Provider's underlying Mongo connection.
func (p *Provider) Close() error {
	p.db = nil
	ctx, cancel := p.connectionContext()
	defer cancel()
	return p.client.Disconnect(ctx)
}

// MoodService returns a MoodService.
func (p *Provider) MoodService() *MoodService { return p.moodService }
