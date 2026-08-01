package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	migrate "github.com/xakep666/mongo-migrate"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		// type defaults to Unspecified on the zero value, but every existing
		// partner predates the Partner/Dealer distinction and was a "partner"
		// in that sense — backfill PartnerTypePartner so they keep showing up
		// as partners rather than as unclassified.
		_, err := db.Collection(collectionName).UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{FieldType: entities.PartnerTypePartner}})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(collectionName).UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{FieldType: ""}})
		return err
	})
}
