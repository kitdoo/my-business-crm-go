package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	migrate "github.com/xakep666/mongo-migrate"
)

func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		// isPublic defaults to false on the zero value, but every existing
		// partner was implicitly public before this field existed (the old
		// ListPublic filter was status=ACTIVE only) — backfill true so the
		// dealer-network listing doesn't silently empty out.
		_, err := db.Collection(collectionName).UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{FieldIsPublic: true}})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(collectionName).UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{FieldIsPublic: ""}})
		return err
	})
}
