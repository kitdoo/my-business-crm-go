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
				Keys: bson.D{{Key: FieldProductID, Value: 1}, {Key: FieldWarehouseID, Value: 1}, {Key: FieldCreatedAt, Value: -1}},
				Options: options.Index().
					SetName("idx_inventory_movements_product_warehouse_created_at"),
			},
			{
				Keys: bson.D{{Key: FieldWarehouseID, Value: 1}, {Key: FieldCreatedAt, Value: -1}},
				Options: options.Index().
					SetName("idx_inventory_movements_warehouse_id_created_at"),
			},
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		if err := coll.Indexes().DropOne(ctx, "idx_inventory_movements_product_warehouse_created_at"); err != nil {
			return err
		}
		return coll.Indexes().DropOne(ctx, "idx_inventory_movements_warehouse_id_created_at")
	})
}
