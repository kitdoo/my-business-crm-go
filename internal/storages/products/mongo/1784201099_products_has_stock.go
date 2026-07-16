package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	migrate "github.com/xakep666/mongo-migrate"
)

// productVariantStatusActive/productSkuStatusActive mirror
// entities.ProductVariantStatusActive/entities.ProductSkuStatusActive
// (int32-aligned with their proto enums) — this migration works against
// the raw driver, one layer below the entities package, so the values are
// inlined rather than imported.
const (
	productVariantStatusActive = int32(2)
	productSkuStatusActive     = int32(2)
)

// backfillHasStock computes entities.Product.HasStock for every currently
// active product via a one-off aggregation joining product_variants,
// product_skus, and inventory — the same three-collection walk
// inventory.Service.ApplyMovement repeats incrementally on every future
// stock movement. Needed once at deploy time since has_stock is a new
// field: existing rows have no value to react from until their next
// movement.
func backfillHasStock(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(collectionName)
	cursor, err := coll.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{FieldDeletedAt: nil}},
		bson.M{"$lookup": bson.M{
			"from": "product_variants",
			"let":  bson.M{"productId": "$" + FieldID},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"$expr": bson.M{"$and": bson.A{
					bson.M{"$eq": bson.A{"$product_id", "$$productId"}},
					bson.M{"$eq": bson.A{"$status", productVariantStatusActive}},
					bson.M{"$eq": bson.A{"$deleted_at", nil}},
				}}}},
				bson.M{"$project": bson.M{"_id": 1}},
			},
			"as": "active_variants",
		}},
		bson.M{"$lookup": bson.M{
			"from": "product_skus",
			"let":  bson.M{"variantIds": "$active_variants._id"},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"$expr": bson.M{"$and": bson.A{
					bson.M{"$in": bson.A{"$variant_id", "$$variantIds"}},
					bson.M{"$eq": bson.A{"$status", productSkuStatusActive}},
					bson.M{"$eq": bson.A{"$deleted_at", nil}},
				}}}},
				bson.M{"$project": bson.M{"_id": 1}},
			},
			"as": "active_skus",
		}},
		bson.M{"$lookup": bson.M{
			"from": "inventory",
			"let":  bson.M{"skuIds": "$active_skus._id"},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"$expr": bson.M{"$and": bson.A{
					bson.M{"$in": bson.A{"$sku_id", "$$skuIds"}},
					bson.M{"$gt": bson.A{"$quantity", 0}},
				}}}},
				bson.M{"$limit": 1},
			},
			"as": "in_stock_rows",
		}},
		bson.M{"$project": bson.M{
			"has_stock": bson.M{"$gt": bson.A{bson.M{"$size": "$in_stock_rows"}, 0}},
		}},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	models := make([]mongo.WriteModel, 0, 256)
	for cursor.Next(ctx) {
		var row struct {
			ID       string `bson:"_id"`
			HasStock bool   `bson:"has_stock"`
		}
		if err := cursor.Decode(&row); err != nil {
			return err
		}
		if !row.HasStock {
			continue // has_stock defaults to false's Go zero value already
		}
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{FieldID: row.ID}).
			SetUpdate(bson.M{"$set": bson.M{FieldHasStock: true}}))
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	if len(models) == 0 {
		return nil
	}
	_, err = coll.BulkWrite(ctx, models)
	return err
}

func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		if _, err := db.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: FieldDeletedAt, Value: 1}, {Key: FieldHasStock, Value: -1}},
			Options: options.Index().
				SetName("idx_products_deleted_at_has_stock"),
		}); err != nil {
			return err
		}
		return backfillHasStock(ctx, db)
	}, func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collectionName)
		if _, err := coll.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{FieldHasStock: false}}); err != nil {
			return err
		}
		return coll.Indexes().DropOne(ctx, "idx_products_deleted_at_has_stock")
	})
}
