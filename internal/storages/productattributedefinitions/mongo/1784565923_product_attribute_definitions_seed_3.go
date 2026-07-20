package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	migrate "github.com/xakep666/mongo-migrate"
)

// seedDefinitionsBatch3 adds "generation" — some catalog items (e.g. stone
// tile products imported from different supplier generations) share a
// display name across a "generation" line and need an explicit,
// admin-set characteristic to tell them apart in the public catalog,
// instead of the web-public frontend guessing it from a category's name
// (fragile: breaks the moment someone renames or restructures categories).
// A follow-up migration, not an edit to the earlier seeds, per this
// collection's "developer adds a migration" convention.
var seedDefinitionsBatch3 = []struct {
	Key      string
	LabelSr  string
	LabelEn  string
	IsPublic bool
	Order    int32
}{
	{Key: "generation", LabelSr: "Generacija", LabelEn: "Generation", IsPublic: true, Order: 11},
}

func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		now := time.Now().UTC()
		docs := make([]any, 0, len(seedDefinitionsBatch3))
		for _, d := range seedDefinitionsBatch3 {
			docs = append(docs, model{
				ID:        bson.NewObjectID().Hex(),
				Key:       d.Key,
				Label:     map[string]string{"sr": d.LabelSr, "en": d.LabelEn},
				IsPublic:  d.IsPublic,
				SortOrder: d.Order,
				CreatedAt: now,
				CursorId:  bson.NewObjectID(),
			})
		}
		_, err := coll.InsertMany(ctx, docs)
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		keys := make([]string, 0, len(seedDefinitionsBatch3))
		for _, d := range seedDefinitionsBatch3 {
			keys = append(keys, d.Key)
		}
		_, err := coll.DeleteMany(ctx, bson.M{FieldKey: bson.M{"$in": keys}})
		return err
	})
}
