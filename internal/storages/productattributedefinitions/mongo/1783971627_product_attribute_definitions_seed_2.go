package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	migrate "github.com/xakep666/mongo-migrate"
)

// seedDefinitionsBatch2 adds the stone/tile-specific characteristics
// (family, standard, format, dimensions) requested after the initial
// catalog — a follow-up migration, not an edit to the first seed, per
// this collection's "developer adds a migration" convention (see
// seedDefinitions in the first seed migration).
var seedDefinitionsBatch2 = []struct {
	Key      string
	LabelSr  string
	LabelEn  string
	IsPublic bool
	Order    int32
}{
	{Key: "family", LabelSr: "Familija", LabelEn: "Family", IsPublic: true, Order: 6},
	{Key: "standard", LabelSr: "Standard", LabelEn: "Standard", IsPublic: true, Order: 7},
	{Key: "format", LabelSr: "Format", LabelEn: "Format", IsPublic: true, Order: 8},
	{Key: "length_mm", LabelSr: "Dužina (mm)", LabelEn: "Length (mm)", IsPublic: true, Order: 9},
	{Key: "width_mm", LabelSr: "Širina (mm)", LabelEn: "Width (mm)", IsPublic: true, Order: 10},
}

func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		now := time.Now().UTC()
		docs := make([]any, 0, len(seedDefinitionsBatch2))
		for _, d := range seedDefinitionsBatch2 {
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
		keys := make([]string, 0, len(seedDefinitionsBatch2))
		for _, d := range seedDefinitionsBatch2 {
			keys = append(keys, d.Key)
		}
		_, err := coll.DeleteMany(ctx, bson.M{FieldKey: bson.M{"$in": keys}})
		return err
	})
}
