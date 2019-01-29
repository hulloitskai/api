package adapter

import (
	"errors"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	mongo "github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	m "github.com/stevenxie/api/pkg/data/mongo"
	"github.com/stevenxie/api/pkg/mood"
	ess "github.com/unixpickle/essentials"
)

// MongoMoodsCollection is the name of Mood objects collection in Mongo.
const MongoMoodsCollection = "moods"

// MongoAdapter implements moods.Repo for a mongo.Database.
type MongoAdapter struct {
	*mongo.Collection
	Config *m.Config
}

// NewMongoAdapter creates a MongoAdapter from an m.DB.
func NewMongoAdapter(db *m.DB) (MongoAdapter, error) {
	if db == nil {
		panic(errors.New("adapter: cannot create MongoAdapter with nil db"))
	}

	// Initialize collection, create 'extId' index.
	var (
		coll        = db.Collection(MongoMoodsCollection)
		ctx, cancel = db.Config.OperationContext()

		unique = true
		model  = mongo.IndexModel{
			Keys:    bson.D{{Key: "extId", Value: -1}},
			Options: &options.IndexOptions{Unique: &unique},
		}
	)
	defer cancel()
	if _, err := coll.Indexes().CreateOne(ctx, model); err != nil {
		return MongoAdapter{}, ess.AddCtx("adapter: creating 'extId' index", err)
	}

	return MongoAdapter{
		Collection: coll,
		Config:     db.Config,
	}, nil
}

// SelectMoods selects the latest `limit` moods from Mongo, starting from the
// record at index `offset`.
func (ma MongoAdapter) SelectMoods(limit int, startID string) ([]*mood.Mood,
	error) {
	ctx, cancel := ma.Config.OperationContext()
	defer cancel()

	// Configure Mongo query.
	var (
		limit64 = int64(limit)
		opts    = options.FindOptions{
			Limit: &limit64,
			Sort:  bson.M{"extId": -1},
		}
		filter = make(bson.M)
	)
	if startID != "" {
		oid, err := primitive.ObjectIDFromHex(startID)
		if err != nil {
			return nil, ess.AddCtx("adapter: parsing hex ID into ObjectID", err)
		}
		filter["_id"] = bson.M{"$gt": oid}
	}

	cur, err := ma.Find(ctx, filter, &opts)
	if err != nil {
		return nil, ess.AddCtx("adapter: performing find query", err)
	}
	defer cur.Close(ctx)

	moods := make([]*mood.Mood, 0, limit)
	for cur.Next(ctx) {
		var result struct {
			mood.Mood `bson:",inline"`
			ID        primitive.ObjectID `bson:"_id"`
		}
		if err = cur.Decode(&result); err != nil {
			return nil, ess.AddCtx("adapter: decoding cursor value", err)
		}
		result.Mood.ID = result.ID.Hex()
		moods = append(moods, &result.Mood)
	}
	if err = cur.Err(); err != nil {
		return nil, ess.AddCtx("adapter: processing results", err)
	}
	if err = cur.Close(ctx); err != nil {
		return nil, ess.AddCtx("adapter: closing cursor", err)
	}
	return moods, nil
}

// InsertMoods inserts moods into Mongo.
func (ma MongoAdapter) InsertMoods(moods []*mood.Mood) error {
	ctx, cancel := ma.Config.OperationContext()
	defer cancel()

	entries := make(bson.A, len(moods))
	for i, mood := range moods {
		data, err := bson.Marshal(mood)
		if err != nil {
			return ess.AddCtx("adapter: encoding mood as BSON", err)
		}
		entries[i] = data
	}

	res, err := ma.InsertMany(ctx, entries)
	if err != nil {
		return ess.AddCtx("adapter: performing insert query", err)
	}

	for i, id := range res.InsertedIDs {
		moods[i].ID = id.(primitive.ObjectID).Hex()
	}
	return nil
}
