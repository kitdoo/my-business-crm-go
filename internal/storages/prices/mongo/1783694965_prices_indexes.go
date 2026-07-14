package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	migrate "github.com/xakep666/mongo-migrate"
)

func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		// Historically accurate as first applied — field renamed
		// product_id -> variant_id afterwards by the
		// 1783980100_backfill_variants migration, which also drops these
		// two indexes and recreates them on the new field/name. Literal
		// strings here (not FieldVariantID/FieldDeletedAt) since this
		// migration's effect, once applied to a database, must never
		// change — only a later migration may un-do or replace it.
		if _, err := db.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			// Partial unique index frees ProductID on soft delete.
			Keys: bson.D{{Key: "product_id", Value: 1}},
			Options: options.Index().
				SetName("idx_product_prices_product_id_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"deleted_at": nil}),
		}); err != nil {
			return err
		}

		_, err := db.Collection(historyCollectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "product_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().
				SetName("idx_product_price_history_product_id_created_at"),
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		if err := db.Collection(collectionName).Indexes().DropOne(ctx, "idx_product_prices_product_id_unique"); err != nil {
			return err
		}
		return db.Collection(historyCollectionName).Indexes().DropOne(ctx, "idx_product_price_history_product_id_created_at")
	})
}
