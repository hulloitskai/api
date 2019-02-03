package mongo

import (
	"fmt"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/stevenxie/api"
	ess "github.com/unixpickle/essentials"
)

// MoodsCollection is the name of Mood objects collection in Mongo.
const MoodsCollection = "moods"

// MoodService implements api.MoodService using a Mongo collection.
type MoodService struct {
	*mongo.Collection
	getContext contextGenerator
}

func newMoodService(db *mongo.Database, ctxgen contextGenerator) (*MoodService,
	error) {
	var (
		unique = true
		model  = mongo.IndexModel{
			Keys:    bson.D{{Key: "extId", Value: -1}},
			Options: &options.IndexOptions{Unique: &unique},
		}
		coll        = db.Collection(MoodsCollection)
		ctx, cancel = ctxgen()
	)
	defer cancel()
	if _, err := coll.Indexes().CreateOne(ctx, model); err != nil {
		return nil, ess.AddCtx("mongo: creating 'extId' index", err)
	}
	return &MoodService{
		Collection: coll,
		getContext: ctxgen,
	}, nil
}

type moodDoc struct {
	api.Mood `bson:",inline"`
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	ExtID    int64              `bson:"extId"`
}

func marshalMoodDoc(src *api.Mood, dst *moodDoc) error {
	dst.Mood = *src
	dst.ExtID = src.ExtID
	if src.ID == "" {
		return nil
	}

	var err error
	dst.ID, err = primitive.ObjectIDFromHex(src.ID)
	return err
}

func unmarshalMoodDoc(src *moodDoc, dst *api.Mood) {
	*dst = src.Mood
	dst.ExtID = src.ExtID
	dst.ID = src.ID.Hex()
}

// CreateMood creates the provided mood, which fills mood.ID.
func (ms *MoodService) CreateMood(mood *api.Mood) error {
	var doc moodDoc
	if err := marshalMoodDoc(mood, &doc); err != nil {
		return ess.AddCtx("mongo: marshalling mood into moodDoc", err)
	}

	// Perform insertion.
	ctx, cancel := ms.getContext()
	defer cancel()
	res, err := ms.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	mood.ID = res.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

// CreateMoods creates the provided moods.
//
// The ID of each api.Mood in moods will be filled with the corresponding
// database ID.
func (ms *MoodService) CreateMoods(moods []*api.Mood) error {
	var (
		docs = make([]interface{}, len(moods))
		err  error
	)
	for i, mood := range moods {
		doc := new(moodDoc)
		if err = marshalMoodDoc(mood, doc); err != nil {
			return ess.AddCtx("mongo: marshalling mood into moodDoc", err)
		}
		docs[i] = doc
	}

	// Perform insertion.
	ctx, cancel := ms.getContext()
	defer cancel()
	res, err := ms.InsertMany(ctx, docs)
	if err != nil {
		return err
	}

	// Check result length.
	if len(res.InsertedIDs) != len(docs) {
		panic(fmt.Errorf("mongo: inserted %d documents, but response contains %d "+
			"IDs", len(docs), len(res.InsertedIDs)))
	}

	for i, id := range res.InsertedIDs {
		moods[i].ID = id.(primitive.ObjectID).Hex()
	}
	return nil
}

// GetMood returns the mood with the specified id.
func (ms *MoodService) GetMood(id string) (*api.Mood, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ess.AddCtx("mongo: converting id into ObjectID", err)
	}

	// Build query.
	filter := bson.M{"_id": oid}

	// Perform query.
	ctx, cancel := ms.getContext()
	defer cancel()
	res := ms.FindOne(ctx, filter)
	if err = res.Err(); err != nil {
		return nil, err
	}

	// Decode result.
	var doc moodDoc
	if err := res.Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, ess.AddCtx("mongo: decoding result as moodDoc", err)
	}

	mood := new(api.Mood)
	unmarshalMoodDoc(&doc, mood)
	return mood, nil
}

// ListMoods lists the last `limit` moods, starting from the mood corresponding
// with `offset`.
func (ms *MoodService) ListMoods(limit, offset int) ([]*api.Mood, error) {
	var (
		limit64 = int64(limit)
		skip64  = int64(offset)
		opts    = options.FindOptions{
			Limit: &limit64,
			Skip:  &skip64,
			Sort:  bson.M{"extId": -1},
		}
	)

	// Perform query.
	ctx, cancel := ms.getContext()
	defer cancel()
	cur, err := ms.Find(ctx, bson.D{}, &opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	// Decode result.
	moods := make([]*api.Mood, 0, limit)
	for cur.Next(ctx) {
		var doc moodDoc
		if err = cur.Decode(&doc); err != nil {
			return nil, ess.AddCtx("mongo: decoding result as moodDoc", err)
		}

		mood := new(api.Mood)
		unmarshalMoodDoc(&doc, mood)
		moods = append(moods, mood)
	}
	if err = cur.Err(); err != nil {
		return nil, ess.AddCtx("mongo: decoding results", err)
	}

	// Close cursor.
	if err = cur.Close(ctx); err != nil {
		return nil, ess.AddCtx("mongo: closing cursor", err)
	}
	return moods, nil
}
