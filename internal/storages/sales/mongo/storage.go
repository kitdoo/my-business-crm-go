package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/altessa-s/go-atlas/domain/converter"

	coreslices "github.com/altessa-s/go-atlas/core/collections/slices"
	coreerrs "github.com/altessa-s/go-atlas/core/errors"
	datamongo "github.com/altessa-s/go-atlas/data/mongo"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	"github.com/kitdoo/my-business-crm-go/internal/storages/sales"
)

const collectionName = "sales"

// countersCollectionName backs Sale.Number — MongoDB has no native
// auto-increment, so each Insert atomically increments a single counter
// document here first (see nextNumber).
const countersCollectionName = "sale_counters"

const (
	FieldID           = "_id"
	FieldNumber       = "number"
	FieldClientID     = "client_id"
	FieldWarehouseID  = "warehouse_id"
	FieldPartnerID    = "partner_id"
	FieldItemProducts = "items.product_id"
	FieldStatus       = "status"
	FieldCreatedBy    = "created_by"
	FieldCreatedAt    = "created_at"
	FieldUpdatedAt    = "updated_at"
	FieldDeletedAt    = "deleted_at"
	FieldEtag         = "etag"
	FieldCursorId     = "cursor_id"
)

const defaultListLimit = datamongo.DefaultListLimit

type itemModel struct {
	ProductID          string `bson:"product_id"`
	Quantity           int64  `bson:"quantity"`
	PriceAmount        int64  `bson:"price_amount"`
	DiscountPercentage int32  `bson:"discount_percentage"`
}

type model struct {
	ID          string              `bson:"_id"`
	Number      int64               `bson:"number"`
	ClientID    string              `bson:"client_id"`
	WarehouseID string              `bson:"warehouse_id"`
	PartnerID   *string             `bson:"partner_id"`
	Items       []itemModel         `bson:"items,omitonupdate"`
	TotalAmount int64               `bson:"total_amount,omitonupdate"`
	Status      entities.SaleStatus `bson:"status"`
	CreatedBy   string              `bson:"created_by,omitonupdate"`
	CreatedAt   time.Time           `bson:"created_at,omitonupdate"`
	UpdatedAt   time.Time           `bson:"updated_at"`
	DeletedAt   *time.Time          `bson:"deleted_at,omitonupdate"`
	Etag        string              `bson:"etag"`
	CursorId    bson.ObjectID       `bson:"cursor_id,omitonupdate"`
}

var _ sales.Storage = (*Storage)(nil)

// Storage implements sales.Storage against MongoDB.
type Storage struct {
	db         *datamongo.Mongo
	collection *mongo.Collection
	counters   *mongo.Collection
}

// New builds a Storage backed by db's "sales" collection.
func New(db *datamongo.Mongo) *Storage {
	return &Storage{
		db:         db,
		collection: db.Database().Collection(collectionName),
		counters:   db.Database().Collection(countersCollectionName),
	}
}

// nextNumber atomically increments and returns the next human-readable
// Sale.Number. $inc + upsert on a single well-known counter document is
// the standard MongoDB auto-increment pattern (no native auto-increment
// exists) — concurrent Inserts each get a distinct, gapless-on-success
// number from this one round trip.
func (s *Storage) nextNumber(ctx context.Context) (int64, error) {
	var doc struct {
		Seq int64 `bson:"seq"`
	}
	err := s.counters.FindOneAndUpdate(
		ctx,
		bson.M{FieldID: collectionName},
		bson.M{"$inc": bson.M{"seq": 1}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&doc)
	if err != nil {
		return 0, coreerrs.WrapOperation(err, "increment sale number counter")
	}
	return doc.Seq, nil
}

func activeOnly(filter bson.M) bson.M {
	filter[FieldDeletedAt] = nil // no Delete path exists for Sale, so this always matches; kept for wire-schema parity
	return filter
}

func (s *Storage) Insert(ctx context.Context, sale *entities.Sale) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	number, err := s.nextNumber(ctx)
	if err != nil {
		return err
	}
	sale.Number = number

	m := converter.Convert(sale, &model{}, converter.WithHandleEmbeddedStructs(true))
	m.CursorId = bson.NewObjectID()
	doc, err := s.db.ConvertToNewDocument(ctx, m)
	if err != nil {
		return coreerrs.WrapOperation(err, "convert sale document")
	}
	if _, err := s.collection.InsertOne(ctx, doc); err != nil {
		return coreerrs.WrapOperation(err, "insert sale")
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, id string) (*entities.Sale, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	var m model
	if err := s.collection.FindOne(ctx, activeOnly(bson.M{FieldID: id})).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrSaleNotFound
		}
		return nil, coreerrs.WrapOperation(err, "get sale")
	}
	return converter.Convert(&m, &entities.Sale{}, converter.WithHandleEmbeddedStructs(true)), nil
}

func (s *Storage) Update(ctx context.Context, sale *entities.Sale, oldEtag string) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	update, err := s.db.ConvertToUpdateDocument(ctx, converter.Convert(sale, &model{}, converter.WithHandleEmbeddedStructs(true)))
	if err != nil {
		return coreerrs.WrapOperation(err, "build update document")
	}
	filter := activeOnly(bson.M{FieldID: sale.ID})
	if oldEtag != "" {
		filter[FieldEtag] = oldEtag
	}
	res, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return coreerrs.WrapOperation(err, "update sale")
	}
	if res.MatchedCount == 0 {
		return errs.ErrStaleEntity
	}
	return nil
}

func (s *Storage) List(ctx context.Context, in *entities.SalesList) (*entities.List[entities.Sale], error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	limit := in.Pagination.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	filter := activeOnly(bson.M{})
	if in.ClientID != nil {
		filter[FieldClientID] = *in.ClientID
	}
	if in.WarehouseID != nil {
		filter[FieldWarehouseID] = *in.WarehouseID
	}
	if in.PartnerID != nil {
		filter[FieldPartnerID] = *in.PartnerID
	}
	if len(in.Statuses) > 0 {
		filter[FieldStatus] = bson.M{"$in": in.Statuses}
	}
	if len(in.CreatedBy) > 0 {
		filter[FieldCreatedBy] = bson.M{"$in": in.CreatedBy}
	}
	if len(in.ProductIDs) > 0 {
		filter[FieldItemProducts] = bson.M{"$in": in.ProductIDs}
	}
	if in.CreatedAt != nil {
		periodFilter(filter, FieldCreatedAt, in.CreatedAt)
	}

	opts := []datamongo.ListCursorOption{
		datamongo.WithListCursorFilter(filter),
		datamongo.WithListCursorSort(listSortField(in.Sort)),
		datamongo.WithListCursorLimit(limit),
		datamongo.WithListCursorTotal(in.IncludeTotalCount),
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
		return nil, coreerrs.WrapOperation(err, "list sales")
	}

	out := &entities.List[entities.Sale]{
		Items: coreslices.To(result.Items, func(m model) *entities.Sale {
			return converter.Convert(&m, &entities.Sale{}, converter.WithHandleEmbeddedStructs(true))
		}),
		Total: result.Total,
	}
	if result.NextCursor != nil {
		out.NextCursor = *result.NextCursor
	}
	return out, nil
}

// listSortField picks the sort key for the requested field. Sale supports
// only created-at sorting (no name).
func listSortField(sort entities.SalesListSort) string {
	field := FieldCreatedAt
	if sort.Direction == entities.SortDirectionDesc {
		return "-" + field
	}
	return field
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
