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
					SetName("idx_products_deleted_at_created_at"),
			},
			{
				Keys: bson.D{{Key: FieldDeletedAt, Value: 1}, {Key: FieldNameSr, Value: 1}},
				Options: options.Index().
					SetName("idx_products_deleted_at_name_sr"),
			},
			{
				Keys: bson.D{{Key: FieldDeletedAt, Value: 1}, {Key: FieldBrandID, Value: 1}},
				Options: options.Index().
					SetName("idx_products_deleted_at_brand_id"),
			},
			{
				Keys: bson.D{{Key: FieldDeletedAt, Value: 1}, {Key: FieldCategoryID, Value: 1}},
				Options: options.Index().
					SetName("idx_products_deleted_at_category_ids"),
			},
			{
				// Partial unique index frees SKU on soft delete.
				Keys: bson.D{{Key: FieldSKU, Value: 1}},
				Options: options.Index().
					SetName("idx_products_sku_unique").
					SetUnique(true).
					SetPartialFilterExpression(bson.M{FieldDeletedAt: nil}),
			},
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		if err := coll.Indexes().DropOne(ctx, "idx_products_deleted_at_created_at"); err != nil {
			return err
		}
		if err := coll.Indexes().DropOne(ctx, "idx_products_deleted_at_name_sr"); err != nil {
			return err
		}
		if err := coll.Indexes().DropOne(ctx, "idx_products_deleted_at_brand_id"); err != nil {
			return err
		}
		if err := coll.Indexes().DropOne(ctx, "idx_products_deleted_at_category_ids"); err != nil {
			return err
		}
		return coll.Indexes().DropOne(ctx, "idx_products_sku_unique")
	})
}
