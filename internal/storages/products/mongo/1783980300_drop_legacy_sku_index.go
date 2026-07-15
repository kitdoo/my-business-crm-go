package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	migrate "github.com/xakep666/mongo-migrate"
)

// legacyProductsSKUUniqueIndex is idx_products_sku_unique, created by
// 1783694352_products_indexes.go back when Product carried sku directly.
// Dropping it was meant to happen in the product_variants backfill
// migration (sku moved off Product onto ProductVariant, then onto
// ProductSKU), but that migration has since been removed. Left in place,
// every product document (which has no sku field at all now) matches the
// partial filter as sku: null, so the *second* Insert ever run collides
// with the first on a duplicate null key.
const legacyProductsSKUUniqueIndex = "idx_products_sku_unique"

// indexNotFoundErrorCode is MongoDB's IndexNotFound command error code.
const indexNotFoundErrorCode = 27

func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		err := db.Collection(collectionName).Indexes().DropOne(ctx, legacyProductsSKUUniqueIndex)
		// Best-effort, like the now-deleted backfill migration that used to
		// drop this same index: an environment where that migration already
		// ran (and already dropped it) has nothing left for this one to do —
		// erroring here would abort every migration behind it and block
		// startup on exactly the environments this fix is meant to help.
		var cmdErr mongo.CommandError
		if errors.As(err, &cmdErr) && cmdErr.HasErrorCode(indexNotFoundErrorCode) {
			return nil
		}
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "sku", Value: 1}},
			Options: options.Index().
				SetName(legacyProductsSKUUniqueIndex).
				SetUnique(true).
				SetPartialFilterExpression(bson.M{FieldDeletedAt: nil}),
		})
		return err
	})
}
