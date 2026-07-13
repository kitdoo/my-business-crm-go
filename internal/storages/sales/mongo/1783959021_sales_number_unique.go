package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	migrate "github.com/xakep666/mongo-migrate"
)

// Belt-and-suspenders on top of the atomic $inc counter (nextNumber):
// that already guarantees Number is unique by construction, this index
// just makes the database reject an insert outright if two concurrent
// Creates ever somehow produced the same number, instead of silently
// persisting a duplicate. No partial filter needed — unlike Client/
// Category, Sale has no soft-delete path (see activeOnly's comment), so
// there's no "freed on delete" case to carve out.
func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: FieldNumber, Value: 1}},
			Options: options.Index().
				SetName("idx_sales_number_unique").
				SetUnique(true),
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		return db.Collection(collectionName).Indexes().DropOne(ctx, "idx_sales_number_unique")
	})
}
