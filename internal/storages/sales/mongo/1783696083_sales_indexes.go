package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	migrate "github.com/xakep666/mongo-migrate"
)

func init() {
	// Historically accurate as first applied — "items.product_id" (literal,
	// not FieldItemVariants) is renamed to items.variant_id afterwards by
	// the 1783980100_backfill_variants migration, which also drops and
	// recreates this index on the new field/name. This migration's effect,
	// once applied to a database, must never change.
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(collectionName).Indexes().CreateMany(ctx, []mongo.IndexModel{
			{
				Keys: bson.D{{Key: FieldDeletedAt, Value: 1}, {Key: FieldCreatedAt, Value: -1}},
				Options: options.Index().
					SetName("idx_sales_deleted_at_created_at"),
			},
			{
				Keys: bson.D{{Key: FieldClientID, Value: 1}},
				Options: options.Index().
					SetName("idx_sales_client_id"),
			},
			{
				Keys: bson.D{{Key: FieldWarehouseID, Value: 1}},
				Options: options.Index().
					SetName("idx_sales_warehouse_id"),
			},
			{
				Keys: bson.D{{Key: FieldPartnerID, Value: 1}},
				Options: options.Index().
					SetName("idx_sales_partner_id"),
			},
			{
				Keys: bson.D{{Key: "items.product_id", Value: 1}},
				Options: options.Index().
					SetName("idx_sales_item_products"),
			},
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		for _, name := range []string{
			"idx_sales_deleted_at_created_at",
			"idx_sales_client_id",
			"idx_sales_warehouse_id",
			"idx_sales_partner_id",
			"idx_sales_item_products",
		} {
			if err := coll.Indexes().DropOne(ctx, name); err != nil {
				return err
			}
		}
		return nil
	})
}
