package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	migrate "github.com/xakep666/mongo-migrate"
)

// Supports the find-or-create-by-email lookup SalesService.Create does
// when a sale carries a new client instead of an existing clientId — not
// unique (email dedup isn't enforced at the storage layer, only the
// find-or-create service logic reuses an existing match), just indexed
// for that lookup.
func init() {
	migrate.MustRegister(func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: FieldDeletedAt, Value: 1}, {Key: FieldEmail, Value: 1}},
			Options: options.Index().
				SetName("idx_clients_deleted_at_email"),
		})
		return err
	}, func(ctx context.Context, db *mongo.Database) error {
		return db.Collection(collectionName).Indexes().DropOne(ctx, "idx_clients_deleted_at_email")
	})
}
