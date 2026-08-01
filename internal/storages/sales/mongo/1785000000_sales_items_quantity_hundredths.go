package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	migrate "github.com/xakep666/mongo-migrate"
)

// Sale.Items[].Quantity switched from whole units to hundredths of a unit
// (the same basis-points convention as priceAmount — see
// entities.SaleItem.Quantity) so fractional quantities (e.g. 2.16 m²) can be
// represented. Existing rows recorded whole units, so every stored value is
// multiplied by 100 here to keep its real meaning under the new scale.
func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(collectionName).UpdateMany(ctx, bson.M{},
			bson.M{"$mul": bson.M{"items.$[].quantity": 100}})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(collectionName).UpdateMany(ctx, bson.M{}, mongo.Pipeline{
			bson.D{{Key: "$set", Value: bson.D{{Key: "items", Value: bson.D{{Key: "$map", Value: bson.D{
				{Key: "input", Value: "$items"},
				{Key: "as", Value: "item"},
				{Key: "in", Value: bson.D{{Key: "$mergeObjects", Value: bson.A{
					"$$item",
					bson.D{{Key: "quantity", Value: bson.D{{Key: "$toLong", Value: bson.D{
						{Key: "$divide", Value: bson.A{"$$item.quantity", 100}},
					}}}}},
				}}}},
			}}}}}}},
		})
		return err
	})
}
