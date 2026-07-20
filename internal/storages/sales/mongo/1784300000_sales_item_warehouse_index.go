package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	migrate "github.com/xakep666/mongo-migrate"
)

// WarehouseID moved off Sale onto each SaleItem (a sale can now draw
// different lines from different warehouses), so the sale-level
// idx_sales_warehouse_id index (1783696083_sales_indexes.go) is replaced
// with one on the nested items.warehouse_id path, mirroring
// idx_sales_item_skus.
func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		if err := coll.Indexes().DropOne(ctx, "idx_sales_warehouse_id"); err != nil {
			return err
		}
		_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: FieldItemWarehouseID, Value: 1}},
			Options: options.Index().
				SetName("idx_sales_item_warehouse_id"),
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		if err := coll.Indexes().DropOne(ctx, "idx_sales_item_warehouse_id"); err != nil {
			return err
		}
		_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "warehouse_id", Value: 1}},
			Options: options.Index().
				SetName("idx_sales_warehouse_id"),
		})
		return err
	})
}
