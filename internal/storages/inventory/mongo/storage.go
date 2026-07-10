package mongo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"

	coreslices "github.com/altessa-s/go-atlas/core/collections/slices"
	coreerrs "github.com/altessa-s/go-atlas/core/errors"
	"github.com/altessa-s/go-atlas/data/cache"
	datamongo "github.com/altessa-s/go-atlas/data/mongo"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	"github.com/kitdoo/my-business-crm-go/internal/storages/inventory"
)

const collectionName = "inventory"

const (
	FieldID          = "_id"
	FieldProductID   = "product_id"
	FieldWarehouseID = "warehouse_id"
	FieldQuantity    = "quantity"
	FieldUpdatedAt   = "updated_at"
	FieldEtag        = "etag"
	FieldCursorId    = "cursor_id"
)

const defaultListLimit = datamongo.DefaultListLimit

type model struct {
	ID          string        `bson:"_id"`
	ProductID   string        `bson:"product_id"`
	WarehouseID string        `bson:"warehouse_id"`
	Quantity    int64         `bson:"quantity"`
	UpdatedAt   time.Time     `bson:"updated_at"`
	Etag        string        `bson:"etag"`
	CursorId    bson.ObjectID `bson:"cursor_id"`
}

var _ inventory.Storage = (*Storage)(nil)

// Storage implements inventory.Storage against MongoDB, with a Redis
// cache-aside layer over Get (see New).
type Storage struct {
	db         *datamongo.Mongo
	collection *mongo.Collection
	cache      *cache.Cache
	logger     *slog.Logger
}

// New builds a Storage backed by db's "inventory" collection. c caches Get
// reads (the TD's stated Redis use case for this hot row); pass
// cache.NewNoop() when Redis is not configured.
func New(db *datamongo.Mongo, c *cache.Cache) *Storage {
	return &Storage{
		db:         db,
		collection: db.Database().Collection(collectionName),
		cache:      c,
		logger:     slog.Default().With(slogx.Module("storage:inventory")),
	}
}

func cacheKey(productID, warehouseID string) string {
	return fmt.Sprintf("inventory:%s:%s", productID, warehouseID)
}

func (s *Storage) Get(ctx context.Context, productID, warehouseID string) (*entities.Inventory, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	var m model
	err := s.cache.GetWithFallback(ctx, cacheKey(productID, warehouseID), &m, func() (any, time.Duration, error) {
		var mm model
		ferr := s.collection.FindOne(ctx, bson.M{FieldProductID: productID, FieldWarehouseID: warehouseID}).Decode(&mm)
		if errors.Is(ferr, mongo.ErrNoDocuments) {
			return nil, 0, cache.ErrMissing
		}
		if ferr != nil {
			return nil, 0, coreerrs.WrapOperation(ferr, "get inventory")
		}
		return &mm, cache.TTLUseDefault, nil
	})
	if err != nil {
		if errors.Is(err, cache.ErrMissing) {
			return nil, errs.ErrInventoryNotFound
		}
		return nil, coreerrs.WrapOperation(err, "get inventory (cache)")
	}
	return converter.Convert(&m, &entities.Inventory{}), nil
}

func (s *Storage) List(ctx context.Context, in *entities.InventoryList) (*entities.List[entities.Inventory], error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	limit := in.Pagination.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	filter := bson.M{}
	if in.ProductID != nil {
		filter[FieldProductID] = *in.ProductID
	}
	if in.WarehouseID != nil {
		filter[FieldWarehouseID] = *in.WarehouseID
	}
	if in.MinQuantity != nil || in.MaxQuantity != nil {
		rng := bson.M{}
		if in.MinQuantity != nil {
			rng["$gte"] = *in.MinQuantity
		}
		if in.MaxQuantity != nil {
			rng["$lte"] = *in.MaxQuantity
		}
		filter[FieldQuantity] = rng
	}

	opts := []datamongo.ListCursorOption{
		datamongo.WithListCursorFilter(filter),
		datamongo.WithListCursorSort(FieldID),
		datamongo.WithListCursorLimit(limit),
	}
	if in.Pagination.Cursor != "" {
		opts = append(opts, datamongo.WithListCursorCursor(in.Pagination.Cursor))
	}

	result, err := datamongo.ListCursor[model](ctx, s.collection, opts...)
	if err != nil {
		if errors.Is(err, datamongo.ErrInvalidCursor) ||
			errors.Is(err, datamongo.ErrCursorChecksumMismatch) ||
			errors.Is(err, datamongo.ErrCursorFilterMismatch) {
			return nil, errs.ErrInvalidListCursor
		}
		return nil, coreerrs.WrapOperation(err, "list inventory")
	}

	out := &entities.List[entities.Inventory]{
		Items: coreslices.To(result.Items, func(m model) *entities.Inventory {
			return converter.Convert(&m, &entities.Inventory{})
		}),
	}
	if result.NextCursor != nil {
		out.NextCursor = *result.NextCursor
	}
	return out, nil
}

// ApplyMovement is the only write path onto Inventory: a single atomic
// FindOneAndUpdate. This deliberately bypasses the standard's "etag/
// timestamp produced in the entity layer" rule — there is no
// load-mutate-save round trip here on purpose, since that would reintroduce
// the read-modify-write race the TD calls out as the reason Inventory
// carries an etag at all. A decrement is additionally guarded by an $expr
// filter so quantity can never go negative; upsert is disabled for
// decrements so a never-seen (productID, warehouseID) pair reports
// errs.ErrInsufficientStock instead of materializing a negative row.
func (s *Storage) ApplyMovement(ctx context.Context, productID, warehouseID string, delta int64) (*entities.Inventory, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	filter := bson.M{FieldProductID: productID, FieldWarehouseID: warehouseID}
	upsert := delta >= 0
	if !upsert {
		filter["$expr"] = bson.M{"$gte": bson.A{bson.M{"$add": bson.A{"$" + FieldQuantity, delta}}, 0}}
	}

	update := bson.M{
		"$inc": bson.M{FieldQuantity: delta},
		"$set": bson.M{FieldUpdatedAt: time.Now().UTC(), FieldEtag: uuid.NewString()},
		"$setOnInsert": bson.M{
			FieldID:       uuid.NewString(),
			FieldCursorId: bson.NewObjectID(),
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(upsert).SetReturnDocument(options.After)
	var m model
	if err := s.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrInsufficientStock
		}
		return nil, coreerrs.WrapOperation(err, "apply inventory movement")
	}

	// Best-effort: the Mongo write already succeeded (the source of truth),
	// so a failed invalidation must not fail the whole movement — that
	// would risk the caller retrying and double-applying delta. Worst case
	// is a stale cached read until the entry's TTL expires.
	if err := s.cache.Delete(ctx, cacheKey(productID, warehouseID)); err != nil {
		s.logger.WarnContext(ctx, "invalidate inventory cache failed",
			slog.String("productId", productID), slog.String("warehouseId", warehouseID), slogx.Error(err))
	}
	return converter.Convert(&m, &entities.Inventory{}), nil
}

func (s *Storage) HasStock(ctx context.Context, warehouseID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	count, err := s.collection.CountDocuments(ctx,
		bson.M{FieldWarehouseID: warehouseID, FieldQuantity: bson.M{"$gt": 0}},
		options.Count().SetLimit(1),
	)
	if err != nil {
		return false, coreerrs.WrapOperation(err, "check warehouse stock")
	}
	return count > 0, nil
}
