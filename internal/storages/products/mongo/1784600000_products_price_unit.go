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
		// priceUnit defaults to Unspecified on the zero value, but every
		// existing product was priced per piece before this field existed
		// (the one known exception, G Facing Brick, gets flipped to
		// PriceUnitSquareMeter by hand after this migration runs) — backfill
		// Piece so the public site doesn't show an unlabeled price for the
		// whole catalog.
		_, err := db.Collection(collectionName).UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{FieldPriceUnit: entities.PriceUnitPiece}})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(collectionName).UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{FieldPriceUnit: ""}})
		return err
	})
}
