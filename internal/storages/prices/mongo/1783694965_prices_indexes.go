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
		if _, err := db.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			// Partial unique index frees SkuID on soft delete.
			Keys: bson.D{{Key: FieldSKUID, Value: 1}},
			Options: options.Index().
				SetName("idx_product_prices_sku_id_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{FieldDeletedAt: nil}),
		}); err != nil {
			return err
		}

		_, err := db.Collection(historyCollectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: FieldSKUID, Value: 1}, {Key: FieldCreatedAt, Value: -1}},
			Options: options.Index().
				SetName("idx_product_price_history_sku_id_created_at"),
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		if err := db.Collection(collectionName).Indexes().DropOne(ctx, "idx_product_prices_sku_id_unique"); err != nil {
			return err
		}
		return db.Collection(historyCollectionName).Indexes().DropOne(ctx, "idx_product_price_history_sku_id_created_at")
	})
}
