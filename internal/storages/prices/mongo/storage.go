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
	"github.com/kitdoo/my-business-crm-go/internal/storages/prices"
)

const (
	collectionName        = "product_prices"
	historyCollectionName = "product_price_history"
)

const (
	FieldID        = "_id"
	FieldProductID = "product_id"
	FieldCreatedAt = "created_at"
	FieldUpdatedAt = "updated_at"
	FieldDeletedAt = "deleted_at"
	FieldEtag      = "etag"
	FieldCursorId  = "cursor_id"
)

const defaultListLimit = datamongo.DefaultListLimit

type model struct {
	ID             string        `bson:"_id"`
	ProductID      string        `bson:"product_id"`
	PriceAmount    int64         `bson:"price_amount"`
	Currency       string        `bson:"currency,omitonupdate"`
	DiscountAmount *int64        `bson:"discount_amount"`
	CreatedAt      time.Time     `bson:"created_at,omitonupdate"`
	UpdatedAt      time.Time     `bson:"updated_at"`
	DeletedAt      *time.Time    `bson:"deleted_at,omitonupdate"`
	Etag           string        `bson:"etag"`
	CursorId       bson.ObjectID `bson:"cursor_id,omitonupdate"`
}

// historyModel has no etag/deleted_at: history entries are immutable
// snapshots, never updated or soft-deleted themselves.
type historyModel struct {
	ID             string        `bson:"_id"`
	ProductID      string        `bson:"product_id"`
	PriceAmount    int64         `bson:"price_amount"`
	Currency       string        `bson:"currency"`
	DiscountAmount *int64        `bson:"discount_amount"`
	CreatedAt      time.Time     `bson:"created_at"`
	CursorId       bson.ObjectID `bson:"cursor_id"`
}

var _ prices.Storage = (*Storage)(nil)

// Storage implements prices.Storage against MongoDB.
type Storage struct {
	db                *datamongo.Mongo
	collection        *mongo.Collection
	historyCollection *mongo.Collection
}

// New builds a Storage backed by db's "product_prices" and
// "product_price_history" collections.
func New(db *datamongo.Mongo) *Storage {
	return &Storage{
		db:                db,
		collection:        db.Database().Collection(collectionName),
		historyCollection: db.Database().Collection(historyCollectionName),
	}
}

func activeOnly(filter bson.M) bson.M {
	filter[FieldDeletedAt] = nil // matches active rows (deleted_at null); aligns with the partial index
	return filter
}

func classifyDuplicate(err error) error {
	isDup, fields := datamongo.IsErrorDuplicate(err)
	if !isDup || fields == nil {
		return nil
	}
	if fields.Contains(FieldProductID) {
		return errs.ErrProductPriceExists
	}
	return nil
}

func (s *Storage) Insert(ctx context.Context, p *entities.ProductPrice) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	m := converter.Convert(p, &model{}, converter.WithHandleEmbeddedStructs(true))
	m.CursorId = bson.NewObjectID()
	doc, err := s.db.ConvertToNewDocument(ctx, m)
	if err != nil {
		return coreerrs.WrapOperation(err, "convert product price document")
	}
	if _, err := s.collection.InsertOne(ctx, doc); err != nil {
		if dupErr := classifyDuplicate(err); dupErr != nil {
			return dupErr
		}
		return coreerrs.WrapOperation(err, "insert product price")
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, id string) (*entities.ProductPrice, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	var m model
	if err := s.collection.FindOne(ctx, activeOnly(bson.M{FieldID: id})).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrProductPriceNotFound
		}
		return nil, coreerrs.WrapOperation(err, "get product price")
	}
	return converter.Convert(&m, &entities.ProductPrice{}), nil
}

func (s *Storage) GetByProductID(ctx context.Context, productID string) (*entities.ProductPrice, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	var m model
	if err := s.collection.FindOne(ctx, activeOnly(bson.M{FieldProductID: productID})).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrProductPriceNotFound
		}
		return nil, coreerrs.WrapOperation(err, "get product price by product")
	}
	return converter.Convert(&m, &entities.ProductPrice{}), nil
}

func (s *Storage) Update(ctx context.Context, p *entities.ProductPrice, oldEtag string) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	update, err := s.db.ConvertToUpdateDocument(ctx, converter.Convert(p, &model{}))
	if err != nil {
		return coreerrs.WrapOperation(err, "build update document")
	}
	filter := activeOnly(bson.M{FieldID: p.ID})
	if oldEtag != "" {
		filter[FieldEtag] = oldEtag
	}
	res, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return coreerrs.WrapOperation(err, "update product price")
	}
	if res.MatchedCount == 0 {
		return errs.ErrStaleEntity
	}
	return nil
}

func (s *Storage) SoftDelete(ctx context.Context, in *entities.SoftDelete) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	filter := activeOnly(bson.M{FieldID: in.ID})
	if in.Etag != "" {
		filter[FieldEtag] = in.Etag
	}
	// Targeted $set: soft delete must not touch sibling fields, so this
	// bypasses ConvertToUpdateDocument (which would require a full model).
	update := bson.M{"$set": bson.M{
		FieldDeletedAt: in.NewUpdatedAt,
		FieldUpdatedAt: in.NewUpdatedAt,
		FieldEtag:      in.NewEtag,
	}}
	res, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return coreerrs.WrapOperation(err, "soft delete product price")
	}
	if res.MatchedCount == 0 {
		if in.Etag != "" {
			return errs.ErrStaleEntity
		}
		return errs.ErrProductPriceNotFound
	}
	return nil
}

func (s *Storage) AppendHistory(ctx context.Context, snapshot *entities.ProductPrice) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	m := converter.Convert(snapshot, &historyModel{}, converter.WithHandleEmbeddedStructs(true))
	m.CursorId = bson.NewObjectID()
	if _, err := s.historyCollection.InsertOne(ctx, m); err != nil {
		return coreerrs.WrapOperation(err, "append product price history")
	}
	return nil
}

func (s *Storage) GetHistory(ctx context.Context, in *entities.ProductPriceGetHistory) (*entities.List[entities.ProductPrice], error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	limit := in.Pagination.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	filter := bson.M{FieldProductID: in.ProductID}
	if in.CreatedAt != nil {
		periodFilter(filter, FieldCreatedAt, in.CreatedAt)
	}

	opts := []datamongo.ListCursorOption{
		datamongo.WithListCursorFilter(filter),
		datamongo.WithListCursorSort("-" + FieldCreatedAt),
		datamongo.WithListCursorLimit(limit),
	}
	if in.Pagination.Cursor != "" {
		opts = append(opts, datamongo.WithListCursorCursor(in.Pagination.Cursor))
	}

	result, err := datamongo.ListCursor[historyModel](ctx, s.historyCollection, opts...)
	if err != nil {
		if errors.Is(err, datamongo.ErrInvalidCursor) ||
			errors.Is(err, datamongo.ErrCursorChecksumMismatch) ||
			errors.Is(err, datamongo.ErrCursorFilterMismatch) {
			return nil, errs.ErrInvalidListCursor
		}
		return nil, coreerrs.WrapOperation(err, "list product price history")
	}

	out := &entities.List[entities.ProductPrice]{
		Items: coreslices.To(result.Items, func(m historyModel) *entities.ProductPrice {
			return converter.Convert(&m, &entities.ProductPrice{})
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
