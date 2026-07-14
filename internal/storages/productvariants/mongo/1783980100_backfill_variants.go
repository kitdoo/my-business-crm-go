package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/google/uuid"

	migrate "github.com/xakep666/mongo-migrate"
)

// Legacy (pre-variant) collection/field/index names, referenced here only —
// the products/prices/inventory/inventorymovements/sales packages no
// longer declare them since sku/imageIds moved off Product onto
// ProductVariant and their FK renamed product_id -> variant_id. All
// migrations run inside one 30s budget shared across every pending
// migration (see go-atlas data/mongo.Mongo.migrate) and a failure here
// rolls *every* migration back — so this uses bulk writes (one round trip
// per collection) instead of one call per product, both for speed and to
// keep this migration from being the one that blows that shared budget.
const (
	legacyProductsCollection     = "products"
	legacyPricesCollection       = "product_prices"
	legacyPriceHistoryCollection = "product_price_history"
	legacyInventoryCollection    = "inventory"
	legacyInvMovementsCollection = "inventory_movements"
	legacySalesCollection        = "sales"
	legacyProductsSKUUniqueIndex = "idx_products_sku_unique"
	legacyPricesUniqueIndex      = "idx_product_prices_product_id_unique"
	legacyPriceHistoryIndex      = "idx_product_price_history_product_id_created_at"
	legacyInventoryUniqueIndex   = "idx_inventory_product_warehouse_unique"
	legacyInvMovementsIndex      = "idx_inventory_movements_product_warehouse_created_at"
	legacySalesItemsIndex        = "idx_sales_item_products"
	newPricesUniqueIndex         = "idx_product_prices_variant_id_unique"
	newPriceHistoryIndex         = "idx_product_price_history_variant_id_created_at"
	newInventoryUniqueIndex      = "idx_inventory_variant_warehouse_unique"
	newInvMovementsIndex         = "idx_inventory_movements_variant_warehouse_created_at"
	newSalesItemsIndex           = "idx_sales_item_variants"
)

// backfillVariantsUp creates one ProductVariant per existing Product
// (copying sku/details/imageIds/status as a starting point — an admin can
// re-split attributes across variants afterwards), then repoints
// ProductPrice/Inventory/InventoryMovement/Sale-item rows from product_id
// to the new variant_id and rebuilds the indexes that were keyed on
// product_id. Product no longer carries sku/imageIds after this point, so
// the old unique SKU index on products is dropped too.
func backfillVariantsUp(ctx context.Context, db *mongo.Database) error {
	products := db.Collection(legacyProductsCollection)
	variants := db.Collection(collectionName)

	cur, err := products.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	productToVariant := make(map[string]string)
	variantDocs := make([]any, 0)

	for cur.Next(ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			return err
		}
		productID, _ := doc["_id"].(string)
		sku, _ := doc["sku"].(string)
		if productID == "" || sku == "" {
			// Already migrated (no legacy sku left) or malformed — skip.
			continue
		}

		var imageIDs []string
		if raw, ok := doc["image_ids"].(bson.A); ok {
			imageIDs = make([]string, 0, len(raw))
			for _, v := range raw {
				if s, ok := v.(string); ok {
					imageIDs = append(imageIDs, s)
				}
			}
		}

		createdAt, _ := doc["created_at"].(time.Time)
		updatedAt, _ := doc["updated_at"].(time.Time)
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		if updatedAt.IsZero() {
			updatedAt = createdAt
		}

		variantID := uuid.NewString()
		variantDocs = append(variantDocs, bson.M{
			FieldID:        variantID,
			FieldProductID: productID,
			FieldSKU:       sku,
			"attributes":   doc["details"],
			"image_ids":    imageIDs,
			FieldStatus:    doc["status"],
			FieldCreatedAt: createdAt,
			FieldUpdatedAt: updatedAt,
			FieldDeletedAt: doc["deleted_at"],
			FieldEtag:      uuid.NewString(),
			FieldCursorId:  bson.NewObjectID(),
		})
		productToVariant[productID] = variantID
	}
	if err := cur.Err(); err != nil {
		return err
	}

	if len(variantDocs) > 0 {
		if _, err := variants.InsertMany(ctx, variantDocs); err != nil {
			return err
		}
	}

	repoint := func(coll *mongo.Collection) error {
		if len(productToVariant) == 0 {
			return nil
		}
		models := make([]mongo.WriteModel, 0, len(productToVariant))
		for productID, variantID := range productToVariant {
			models = append(models, mongo.NewUpdateManyModel().
				SetFilter(bson.M{"product_id": productID}).
				SetUpdate(bson.M{"$set": bson.M{"variant_id": variantID}, "$unset": bson.M{"product_id": ""}}))
		}
		_, err := coll.BulkWrite(ctx, models)
		return err
	}
	if err := repoint(db.Collection(legacyPricesCollection)); err != nil {
		return err
	}
	if err := repoint(db.Collection(legacyPriceHistoryCollection)); err != nil {
		return err
	}
	if err := repoint(db.Collection(legacyInventoryCollection)); err != nil {
		return err
	}
	if err := repoint(db.Collection(legacyInvMovementsCollection)); err != nil {
		return err
	}

	// Sale items are a nested array (items.$[].product_id), not a
	// top-level field — arrayFilters targets every element needing a
	// rewrite regardless of position.
	if len(productToVariant) > 0 {
		sales := db.Collection(legacySalesCollection)
		models := make([]mongo.WriteModel, 0, len(productToVariant))
		for productID, variantID := range productToVariant {
			models = append(models, mongo.NewUpdateManyModel().
				SetFilter(bson.M{"items.product_id": productID}).
				SetUpdate(bson.M{"$set": bson.M{"items.$[elem].variant_id": variantID}, "$unset": bson.M{"items.$[elem].product_id": ""}}).
				SetArrayFilters([]any{bson.M{"elem.product_id": productID}}))
		}
		if _, err := sales.BulkWrite(ctx, models); err != nil {
			return err
		}
	}

	// Rebuild the indexes that were keyed on product_id — best-effort:
	// DropOne on an index that never existed on a database provisioned
	// fresh after this point returns an error that's fine to ignore here.
	_ = products.Indexes().DropOne(ctx, legacyProductsSKUUniqueIndex)

	if err := renameIndex(ctx, db.Collection(legacyPricesCollection), legacyPricesUniqueIndex,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "variant_id", Value: 1}},
			Options: options.Index().SetName(newPricesUniqueIndex).SetUnique(true).SetPartialFilterExpression(bson.M{"deleted_at": nil}),
		}); err != nil {
		return err
	}
	if err := renameIndex(ctx, db.Collection(legacyPriceHistoryCollection), legacyPriceHistoryIndex,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "variant_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName(newPriceHistoryIndex),
		}); err != nil {
		return err
	}
	if err := renameIndex(ctx, db.Collection(legacyInventoryCollection), legacyInventoryUniqueIndex,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "variant_id", Value: 1}, {Key: "warehouse_id", Value: 1}},
			Options: options.Index().SetName(newInventoryUniqueIndex).SetUnique(true),
		}); err != nil {
		return err
	}
	if err := renameIndex(ctx, db.Collection(legacyInvMovementsCollection), legacyInvMovementsIndex,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "variant_id", Value: 1}, {Key: "warehouse_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName(newInvMovementsIndex),
		}); err != nil {
		return err
	}
	if err := renameIndex(ctx, db.Collection(legacySalesCollection), legacySalesItemsIndex,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "items.variant_id", Value: 1}},
			Options: options.Index().SetName(newSalesItemsIndex),
		}); err != nil {
		return err
	}

	return nil
}

// renameIndex drops oldName (best-effort — absent on a database
// provisioned fresh after this point) and creates newModel in its place.
func renameIndex(ctx context.Context, coll *mongo.Collection, oldName string, newModel mongo.IndexModel) error {
	_ = coll.Indexes().DropOne(ctx, oldName)
	_, err := coll.Indexes().CreateOne(ctx, newModel)
	return err
}

// backfillVariantsDown reverses backfillVariantsUp on a best-effort basis
// (dev/rollback use, not a guaranteed lossless prod downgrade): restores
// sku/imageIds onto Product, repoints variant_id back to product_id
// (including sale items), drops every ProductVariant row, and rebuilds the
// original product_id-keyed indexes.
func backfillVariantsDown(ctx context.Context, db *mongo.Database) error {
	products := db.Collection(legacyProductsCollection)
	variants := db.Collection(collectionName)

	cur, err := variants.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	variantToProduct := make(map[string]string)
	productModels := make([]mongo.WriteModel, 0)

	for cur.Next(ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			return err
		}
		variantID, _ := doc["_id"].(string)
		productID, _ := doc["product_id"].(string)
		if variantID == "" || productID == "" {
			continue
		}
		variantToProduct[variantID] = productID

		productModels = append(productModels, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": productID}).
			SetUpdate(bson.M{"$set": bson.M{"sku": doc["sku"], "image_ids": doc["image_ids"]}}))
	}
	if err := cur.Err(); err != nil {
		return err
	}
	if len(productModels) > 0 {
		if _, err := products.BulkWrite(ctx, productModels); err != nil {
			return err
		}
	}

	repoint := func(coll *mongo.Collection) error {
		if len(variantToProduct) == 0 {
			return nil
		}
		models := make([]mongo.WriteModel, 0, len(variantToProduct))
		for variantID, productID := range variantToProduct {
			models = append(models, mongo.NewUpdateManyModel().
				SetFilter(bson.M{"variant_id": variantID}).
				SetUpdate(bson.M{"$set": bson.M{"product_id": productID}, "$unset": bson.M{"variant_id": ""}}))
		}
		_, err := coll.BulkWrite(ctx, models)
		return err
	}
	if err := repoint(db.Collection(legacyPricesCollection)); err != nil {
		return err
	}
	if err := repoint(db.Collection(legacyPriceHistoryCollection)); err != nil {
		return err
	}
	if err := repoint(db.Collection(legacyInventoryCollection)); err != nil {
		return err
	}
	if err := repoint(db.Collection(legacyInvMovementsCollection)); err != nil {
		return err
	}

	if len(variantToProduct) > 0 {
		sales := db.Collection(legacySalesCollection)
		models := make([]mongo.WriteModel, 0, len(variantToProduct))
		for variantID, productID := range variantToProduct {
			models = append(models, mongo.NewUpdateManyModel().
				SetFilter(bson.M{"items.variant_id": variantID}).
				SetUpdate(bson.M{"$set": bson.M{"items.$[elem].product_id": productID}, "$unset": bson.M{"items.$[elem].variant_id": ""}}).
				SetArrayFilters([]any{bson.M{"elem.variant_id": variantID}}))
		}
		if _, err := sales.BulkWrite(ctx, models); err != nil {
			return err
		}
	}

	if _, err := variants.DeleteMany(ctx, bson.M{}); err != nil {
		return err
	}

	if err := renameIndex(ctx, db.Collection(legacyPricesCollection), newPricesUniqueIndex,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "product_id", Value: 1}},
			Options: options.Index().SetName(legacyPricesUniqueIndex).SetUnique(true).SetPartialFilterExpression(bson.M{"deleted_at": nil}),
		}); err != nil {
		return err
	}
	if err := renameIndex(ctx, db.Collection(legacyPriceHistoryCollection), newPriceHistoryIndex,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "product_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName(legacyPriceHistoryIndex),
		}); err != nil {
		return err
	}
	if err := renameIndex(ctx, db.Collection(legacyInventoryCollection), newInventoryUniqueIndex,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "product_id", Value: 1}, {Key: "warehouse_id", Value: 1}},
			Options: options.Index().SetName(legacyInventoryUniqueIndex).SetUnique(true),
		}); err != nil {
		return err
	}
	if err := renameIndex(ctx, db.Collection(legacyInvMovementsCollection), newInvMovementsIndex,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "product_id", Value: 1}, {Key: "warehouse_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName(legacyInvMovementsIndex),
		}); err != nil {
		return err
	}
	if err := renameIndex(ctx, db.Collection(legacySalesCollection), newSalesItemsIndex,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "items.product_id", Value: 1}},
			Options: options.Index().SetName(legacySalesItemsIndex),
		}); err != nil {
		return err
	}

	return nil
}

func init() {
	migrate.MustRegister(backfillVariantsUp, backfillVariantsDown)
}
