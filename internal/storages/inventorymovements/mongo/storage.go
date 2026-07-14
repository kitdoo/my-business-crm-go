package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/altessa-s/go-atlas/domain/converter"

	coreslices "github.com/altessa-s/go-atlas/core/collections/slices"
	coreerrs "github.com/altessa-s/go-atlas/core/errors"
	datamongo "github.com/altessa-s/go-atlas/data/mongo"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	"github.com/kitdoo/my-business-crm-go/internal/storages/inventorymovements"
)

const collectionName = "inventory_movements"

const (
	FieldID          = "_id"
	FieldVariantID   = "variant_id"
	FieldWarehouseID = "warehouse_id"
	FieldType        = "type"
	FieldCreatedBy   = "created_by"
	FieldCreatedAt   = "created_at"
	FieldCursorId    = "cursor_id"
)

const defaultListLimit = datamongo.DefaultListLimit

type model struct {
	ID          string                `bson:"_id"`
	VariantID   string                `bson:"variant_id"`
	WarehouseID string                `bson:"warehouse_id"`
	Type        entities.MovementType `bson:"type"`
	Quantity    int64                 `bson:"quantity"`
	Comment     string                `bson:"comment"`
	SaleID      *string               `bson:"sale_id"`
	CreatedBy   string                `bson:"created_by"`
	CreatedAt   time.Time             `bson:"created_at"`
	CursorId    bson.ObjectID         `bson:"cursor_id"`
}

var _ inventorymovements.Storage = (*Storage)(nil)

// Storage implements inventorymovements.Storage against MongoDB.
type Storage struct {
	db         *datamongo.Mongo
	collection *mongo.Collection
}

// New builds a Storage backed by db's "inventory_movements" collection.
func New(db *datamongo.Mongo) *Storage {
	return &Storage{
		db:         db,
		collection: db.Database().Collection(collectionName),
	}
}

func (s *Storage) Insert(ctx context.Context, m *entities.InventoryMovement) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	doc := converter.Convert(m, &model{}, converter.WithHandleEmbeddedStructs(true))
	doc.CursorId = bson.NewObjectID()
	if _, err := s.collection.InsertOne(ctx, doc); err != nil {
		return coreerrs.WrapOperation(err, "insert inventory movement")
	}
	return nil
}

func (s *Storage) List(ctx context.Context, in *entities.InventoryMovementsList) (*entities.List[entities.InventoryMovement], error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	filter := bson.M{}
	if in.WarehouseID != nil {
		filter[FieldWarehouseID] = *in.WarehouseID
	}
	if len(in.Types) > 0 {
		filter[FieldType] = bson.M{"$in": in.Types}
	}
	if len(in.VariantIDs) > 0 {
		filter[FieldVariantID] = bson.M{"$in": in.VariantIDs}
	}
	if len(in.CreatedBy) > 0 {
		filter[FieldCreatedBy] = bson.M{"$in": in.CreatedBy}
	}
	if in.CreatedAt != nil {
		periodFilter(filter, FieldCreatedAt, in.CreatedAt)
	}

	return s.list(ctx, filter, in.Pagination)
}

func (s *Storage) GetHistory(ctx context.Context, in *entities.InventoryMovementGetHistory) (*entities.List[entities.InventoryMovement], error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	filter := bson.M{FieldVariantID: in.VariantID, FieldWarehouseID: in.WarehouseID}
	if len(in.Types) > 0 {
		filter[FieldType] = bson.M{"$in": in.Types}
	}
	if in.CreatedAt != nil {
		periodFilter(filter, FieldCreatedAt, in.CreatedAt)
	}

	return s.list(ctx, filter, in.Pagination)
}

func (s *Storage) list(ctx context.Context, filter bson.M, pagination entities.ListPagination) (*entities.List[entities.InventoryMovement], error) {
	limit := pagination.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	opts := []datamongo.ListCursorOption{
		datamongo.WithListCursorFilter(filter),
		datamongo.WithListCursorSort("-" + FieldCreatedAt),
		datamongo.WithListCursorLimit(limit),
	}
	if pagination.Cursor != "" {
		opts = append(opts, datamongo.WithListCursorCursor(pagination.Cursor))
	}

	result, err := datamongo.ListCursor[model](ctx, s.collection, opts...)
	if err != nil {
		if errors.Is(err, datamongo.ErrInvalidCursor) ||
			errors.Is(err, datamongo.ErrCursorChecksumMismatch) ||
			errors.Is(err, datamongo.ErrCursorFilterMismatch) {
			return nil, errs.ErrInvalidListCursor
		}
		return nil, coreerrs.WrapOperation(err, "list inventory movements")
	}

	out := &entities.List[entities.InventoryMovement]{
		Items: coreslices.To(result.Items, func(m model) *entities.InventoryMovement {
			return converter.Convert(&m, &entities.InventoryMovement{})
		}),
		Total: result.Total,
	}
	if result.NextCursor != nil {
		out.NextCursor = *result.NextCursor
	}
	return out, nil
}

func periodFilter(filter bson.M, field string, period *entities.PeriodFilter) {
	rng := bson.M{}
	if period.From != nil {
		rng["$gte"] = *period.From
	}
	if period.To != nil {
		rng["$lte"] = *period.To
	}
	if len(rng) > 0 {
		filter[field] = rng
	}
}
