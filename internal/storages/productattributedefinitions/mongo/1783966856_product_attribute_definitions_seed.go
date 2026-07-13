package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	migrate "github.com/xakep666/mongo-migrate"
)

// seedDefinitions is the initial characteristics catalog. Add more via a
// follow-up migration when a new characteristic is needed — there is no
// admin UI for this, it's managed directly in the database by a developer.
var seedDefinitions = []struct {
	Key      string
	LabelSr  string
	LabelEn  string
	IsPublic bool
	Order    int32
}{
	{Key: "material", LabelSr: "Materijal", LabelEn: "Material", IsPublic: true, Order: 1},
	{Key: "color", LabelSr: "Boja", LabelEn: "Color", IsPublic: true, Order: 2},
	{Key: "size", LabelSr: "Dimenzije", LabelEn: "Dimensions", IsPublic: true, Order: 3},
	{Key: "weight", LabelSr: "Težina", LabelEn: "Weight", IsPublic: true, Order: 4},
	{Key: "supplierNote", LabelSr: "Napomena dobavljača", LabelEn: "Supplier note", IsPublic: false, Order: 5},
}

func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		if _, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
			{
				Keys:    bson.D{{Key: FieldKey, Value: 1}},
				Options: options.Index().SetName("idx_product_attribute_definitions_key_unique").SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: FieldSortOrder, Value: 1}},
				Options: options.Index().SetName("idx_product_attribute_definitions_sort_order"),
			},
		}); err != nil {
			return err
		}

		now := time.Now().UTC()
		docs := make([]any, 0, len(seedDefinitions))
		for _, d := range seedDefinitions {
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
		keys := make([]string, 0, len(seedDefinitions))
		for _, d := range seedDefinitions {
			keys = append(keys, d.Key)
		}
		if _, err := coll.DeleteMany(ctx, bson.M{FieldKey: bson.M{"$in": keys}}); err != nil {
			return err
		}
		if err := coll.Indexes().DropOne(ctx, "idx_product_attribute_definitions_key_unique"); err != nil {
			return err
		}
		return coll.Indexes().DropOne(ctx, "idx_product_attribute_definitions_sort_order")
	})
}
