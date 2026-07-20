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
				// Historical: sale-level warehouse_id, since removed —
				// literal key kept as-is rather than referencing a current
				// constant, see 1784300000_sales_item_warehouse_index.go
				// which drops this index and replaces it with one on
				// items.warehouse_id.
				Keys: bson.D{{Key: "warehouse_id", Value: 1}},
				Options: options.Index().
					SetName("idx_sales_warehouse_id"),
			},
			{
				Keys: bson.D{{Key: FieldPartnerID, Value: 1}},
				Options: options.Index().
					SetName("idx_sales_partner_id"),
			},
			{
				Keys: bson.D{{Key: "items.sku_id", Value: 1}},
				Options: options.Index().
					SetName("idx_sales_item_skus"),
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
			"idx_sales_item_skus",
		} {
			if err := coll.Indexes().DropOne(ctx, name); err != nil {
				return err
			}
		}
		return nil
	})
}
