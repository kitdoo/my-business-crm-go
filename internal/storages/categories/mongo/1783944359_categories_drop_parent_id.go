package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	migrate "github.com/xakep666/mongo-migrate"
)

// Categories dropped the parentId concept entirely (they're peers, never a
// hierarchy) — this removes the now-unused index and field left over from
// when Category still carried a self-referencing parent_id.
func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		if err := coll.Indexes().DropOne(ctx, "idx_categories_deleted_at_parent_id"); err != nil {
			return err
		}
		_, err := coll.UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"parent_id": ""}})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		// Best-effort only: the parent_id values themselves are gone for
		// good once $unset above has run (never persisted anywhere to
		// restore from) — this just recreates the index shape so a
		// rollback leaves the collection schema-compatible with the prior
		// migration, not the data.
		_, err := db.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: FieldDeletedAt, Value: 1}, {Key: "parent_id", Value: 1}},
			Options: options.Index().
				SetName("idx_categories_deleted_at_parent_id"),
		})
		return err
	})
}
