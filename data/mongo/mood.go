package mongo

import (
	"context"
	"time"

	errors "golang.org/x/xerrors"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"

	"github.com/stevenxie/api"
	"github.com/stevenxie/api/internal/util"
)

// MoodsCollection is the name of Mood objects collection in Mongo.
const MoodsCollection = "moods"

// A MoodService is an implementation of api.MoodService that uses Mongo as the
// underlying data store.
type MoodService struct {
	coll    *mongo.Collection
	timeout time.Duration
}

func newMoodService(db *mongo.Database) *MoodService {
	return &MoodService{coll: db.Collection(MoodsCollection)}
}

// SetTimeout sets the timeout for a MongoDB operation.
func (ms *MoodService) SetTimeout(timeout time.Duration) {
	ms.timeout = timeout
}

// init initializes the moods collection, and associated indexes.
func (ms *MoodService) init() error {
	var (
		unique = true
		model  = mongo.IndexModel{
			Keys:    bson.D{{Key: "extId", Value: -1}},
			Options: &options.IndexOptions{Unique: &unique},
		}
		ctx, cancel = ms.context()
	)
	defer cancel()
	if _, err := ms.coll.Indexes().CreateOne(ctx, model); err != nil {
		return errors.Errorf("mongo: creating 'extId' index: %w", err)
	}
	return nil
}

func (ms *MoodService) context() (context.Context, context.CancelFunc) {
	return util.ContextWithTimeout(ms.timeout)
}

// CreateMood creates the provided mood, which fills mood.ID.
func (ms *MoodService) CreateMood(mood *api.Mood) error {
	var doc moodDoc
	if err := marshalMoodDoc(mood, &doc); err != nil {
		return errors.Errorf("mongo: marshalling mood into moodDoc: %w", err)
	}

	// Perform insertion.
	ctx, cancel := ms.context()
	defer cancel()
	res, err := ms.coll.InsertOne(ctx, doc)
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
			return errors.Errorf("mongo: marshalling mood into moodDoc: %w", err)
		}
		docs[i] = doc
	}

	// Perform insertion.
	ctx, cancel := ms.context()
	defer cancel()
	res, err := ms.coll.InsertMany(ctx, docs)
	if err != nil {
		return err
	}

	// Check result length.
	if len(res.InsertedIDs) != len(docs) {
		panic(errors.Errorf("mongo: inserted %d documents, but response "+
			"contains %d IDs", len(docs), len(res.InsertedIDs)))
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
		return nil, util.WrapCause(ErrInvalidID, err)
	}

	// Build query.
	filter := bson.M{"_id": oid}

	// Perform query.
	ctx, cancel := ms.context()
	defer cancel()
	res := ms.coll.FindOne(ctx, filter)
	if err = res.Err(); err != nil {
		return nil, err
	}

	// Decode result.
	var doc moodDoc
	if err := res.Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.Errorf("mongo: decoding result as moodDoc: %w", err)
	}

	mood := new(api.Mood)
	unmarshalMoodDoc(&doc, mood)
	return mood, nil
}

// ListMoods lists the last `limit` moods, starting from the mood corresponding
// with `offset`.
func (ms *MoodService) ListMoods(limit, offset int) ([]*api.Mood, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"extId": -1})

	// Perform query.
	ctx, cancel := ms.context()
	defer cancel()
	cur, err := ms.coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	// Decode result.
	moods := make([]*api.Mood, 0, limit)
	for cur.Next(ctx) {
		var doc moodDoc
		if err = cur.Decode(&doc); err != nil {
			return nil, errors.Errorf("mongo: decoding result as moodDoc: %w", err)
		}

		mood := new(api.Mood)
		unmarshalMoodDoc(&doc, mood)
		moods = append(moods, mood)
	}
	if err = cur.Err(); err != nil {
		return nil, errors.Errorf("mongo: decoding results: %w", err)
	}

	// Close cursor.
	if err = cur.Close(ctx); err != nil {
		return nil, errors.Errorf("mongo: closing cursor: %w", err)
	}
	return moods, nil
}

// DeleteMood deletes the mood with the specified id.
func (ms *MoodService) DeleteMood(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return util.WrapCause(ErrInvalidID, err)
	}

	// Build query.
	filter := bson.M{"_id": oid}

	// Perform query.
	ctx, cancel := ms.context()
	defer cancel()
	res := ms.coll.FindOneAndDelete(ctx, filter)
	return res.Err()
}

type moodDoc struct {
	api.Mood `bson:",inline"`
	OID      primitive.ObjectID `bson:"_id,omitempty"`
	ExtID    int64              `bson:"extId"`
}

func marshalMoodDoc(src *api.Mood, dst *moodDoc) error {
	dst.Mood = *src
	dst.ExtID = src.ExtID
	if src.ID == "" {
		return nil
	}

	var err error
	dst.OID, err = primitive.ObjectIDFromHex(src.ID)
	return err
}

func unmarshalMoodDoc(src *moodDoc, dst *api.Mood) {
	*dst = src.Mood
	dst.ExtID = src.ExtID

	if len(src.OID) > 0 {
		dst.ID = src.OID.Hex()
	}
}
