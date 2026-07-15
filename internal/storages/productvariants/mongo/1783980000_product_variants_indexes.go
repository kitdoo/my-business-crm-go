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
					SetName("idx_product_variants_deleted_at_created_at"),
			},
			{
				Keys: bson.D{{Key: FieldDeletedAt, Value: 1}, {Key: FieldProductID, Value: 1}},
				Options: options.Index().
					SetName("idx_product_variants_deleted_at_product_id"),
			},
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		if err := coll.Indexes().DropOne(ctx, "idx_product_variants_deleted_at_created_at"); err != nil {
			return err
		}
		return coll.Indexes().DropOne(ctx, "idx_product_variants_deleted_at_product_id")
	})
}
