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
					SetName("idx_categories_deleted_at_created_at"),
			},
			{
				Keys: bson.D{{Key: FieldDeletedAt, Value: 1}, {Key: FieldNameSr, Value: 1}},
				Options: options.Index().
					SetName("idx_categories_deleted_at_name_sr"),
			},
			{
				// Literal "parent_id", not a shared FieldParentID const —
				// the field itself was removed from the Category type (see
				// the later migration that drops this index); this
				// historical migration is left as a record of what ran,
				// just no longer referencing app code that no longer exists.
				Keys: bson.D{{Key: FieldDeletedAt, Value: 1}, {Key: "parent_id", Value: 1}},
				Options: options.Index().
					SetName("idx_categories_deleted_at_parent_id"),
			},
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		if err := coll.Indexes().DropOne(ctx, "idx_categories_deleted_at_created_at"); err != nil {
			return err
		}
		if err := coll.Indexes().DropOne(ctx, "idx_categories_deleted_at_name_sr"); err != nil {
			return err
		}
		return coll.Indexes().DropOne(ctx, "idx_categories_deleted_at_parent_id")
	})
}
