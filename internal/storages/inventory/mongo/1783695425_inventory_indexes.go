package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	migrate "github.com/xakep666/mongo-migrate"
)

func init() {
	// Historically accurate as first applied — the "product_id" field key
	// (literal, not FieldVariantID) is renamed to variant_id afterwards by
	// the 1783980100_backfill_variants migration, which also drops and
	// recreates these indexes on the new field/name. This migration's
	// effect, once applied to a database, must never change.
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(collectionName).Indexes().CreateMany(ctx, []mongo.IndexModel{
			{
				// Unique per (productId, warehouseId), per the TD.
				Keys: bson.D{{Key: "product_id", Value: 1}, {Key: FieldWarehouseID, Value: 1}},
				Options: options.Index().
					SetName("idx_inventory_product_warehouse_unique").
					SetUnique(true),
			},
			{
				Keys: bson.D{{Key: FieldWarehouseID, Value: 1}, {Key: FieldQuantity, Value: 1}},
				Options: options.Index().
					SetName("idx_inventory_warehouse_id_quantity"),
			},
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		if err := coll.Indexes().DropOne(ctx, "idx_inventory_product_warehouse_unique"); err != nil {
			return err
		}
		return coll.Indexes().DropOne(ctx, "idx_inventory_warehouse_id_quantity")
	})
}
