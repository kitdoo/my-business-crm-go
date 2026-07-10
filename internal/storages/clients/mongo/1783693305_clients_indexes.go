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
					SetName("idx_clients_deleted_at_created_at"),
			},
			{
				Keys: bson.D{{Key: FieldDeletedAt, Value: 1}, {Key: FieldName, Value: 1}},
				Options: options.Index().
					SetName("idx_clients_deleted_at_name"),
			},
			{
				// Partial unique index frees Phone on soft delete.
				Keys: bson.D{{Key: FieldPhone, Value: 1}},
				Options: options.Index().
					SetName("idx_clients_phone_unique").
					SetUnique(true).
					SetPartialFilterExpression(bson.M{FieldDeletedAt: nil}),
			},
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		if err := coll.Indexes().DropOne(ctx, "idx_clients_deleted_at_created_at"); err != nil {
			return err
		}
		if err := coll.Indexes().DropOne(ctx, "idx_clients_deleted_at_name"); err != nil {
			return err
		}
		return coll.Indexes().DropOne(ctx, "idx_clients_phone_unique")
	})
}
