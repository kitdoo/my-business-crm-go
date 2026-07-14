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
	"github.com/kitdoo/my-business-crm-go/internal/storages/products"
)

const collectionName = "products"

const (
	FieldID = "_id"
	// FieldNameSr is the sort key for Name. "sr" is the mandatory locale
	// (see common_localized_string.proto), used as a deterministic single
	// key since Name itself is a multi-locale map.
	FieldNameSr     = "name.sr"
	FieldBrandID    = "brand_id"
	FieldCategoryID = "category_ids"
	FieldStatus     = "status"
	FieldCreatedAt  = "created_at"
	FieldUpdatedAt  = "updated_at"
	FieldDeletedAt  = "deleted_at"
	FieldEtag       = "etag"
	FieldCursorId   = "cursor_id"
)

const defaultListLimit = datamongo.DefaultListLimit

type model struct {
	ID          string                              `bson:"_id"`
	Name        entities.LocalizedString            `bson:"name"`
	Description entities.LocalizedString            `bson:"description"`
	BrandID     string                              `bson:"brand_id"`
	CategoryIDs []string                            `bson:"category_ids"`
	Details     map[string]entities.LocalizedString `bson:"details"`
	Status      entities.ProductStatus              `bson:"status"`
	CreatedAt   time.Time                           `bson:"created_at,omitonupdate"`
	UpdatedAt   time.Time                           `bson:"updated_at"`
	DeletedAt   *time.Time                          `bson:"deleted_at,omitonupdate"`
	Etag        string                              `bson:"etag"`
	CursorId    bson.ObjectID                       `bson:"cursor_id,omitonupdate"`
}

var _ products.Storage = (*Storage)(nil)

// Storage implements products.Storage against MongoDB.
type Storage struct {
	db         *datamongo.Mongo
	collection *mongo.Collection
}

// New builds a Storage backed by db's "products" collection.
func New(db *datamongo.Mongo) *Storage {
	return &Storage{
		db:         db,
		collection: db.Database().Collection(collectionName),
	}
}

func activeOnly(filter bson.M) bson.M {
	filter[FieldDeletedAt] = nil // matches active rows (deleted_at null); aligns with the partial index
	return filter
}

func (s *Storage) Insert(ctx context.Context, p *entities.Product) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	m := converter.Convert(p, &model{}, converter.WithHandleEmbeddedStructs(true))
	m.CursorId = bson.NewObjectID()
	doc, err := s.db.ConvertToNewDocument(ctx, m)
	if err != nil {
		return coreerrs.WrapOperation(err, "convert product document")
	}
	if _, err := s.collection.InsertOne(ctx, doc); err != nil {
		return coreerrs.WrapOperation(err, "insert product")
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, id string) (*entities.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	var m model
	if err := s.collection.FindOne(ctx, activeOnly(bson.M{FieldID: id})).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrProductNotFound
		}
		return nil, coreerrs.WrapOperation(err, "get product")
	}
	return converter.Convert(&m, &entities.Product{}), nil
}

func (s *Storage) Update(ctx context.Context, p *entities.Product, oldEtag string) error {
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
		return coreerrs.WrapOperation(err, "update product")
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
		return coreerrs.WrapOperation(err, "soft delete product")
	}
	if res.MatchedCount == 0 {
		if in.Etag != "" {
			return errs.ErrStaleEntity
		}
		return errs.ErrProductNotFound
	}
	return nil
}

func (s *Storage) List(ctx context.Context, in *entities.ProductsList) (*entities.List[entities.Product], error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	limit := in.Pagination.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	filter := activeOnly(bson.M{})
	if in.BrandID != nil {
		filter[FieldBrandID] = *in.BrandID
	}
	if in.CategoryID != nil {
		filter[FieldCategoryID] = *in.CategoryID
	}
	if len(in.Statuses) > 0 {
		filter[FieldStatus] = bson.M{"$in": in.Statuses}
	}
	if len(in.BrandIDs) > 0 {
		filter[FieldBrandID] = bson.M{"$in": in.BrandIDs}
	}
	if len(in.CategoryIDs) > 0 {
		filter[FieldCategoryID] = bson.M{"$in": in.CategoryIDs}
	}
	if len(in.IDs) > 0 {
		filter[FieldID] = bson.M{"$in": in.IDs}
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
		return nil, coreerrs.WrapOperation(err, "list products")
	}

	out := &entities.List[entities.Product]{
		Items: coreslices.To(result.Items, func(m model) *entities.Product {
			return converter.Convert(&m, &entities.Product{})
		}),
		Total: result.Total,
	}
	if result.NextCursor != nil {
		out.NextCursor = *result.NextCursor
	}
	return out, nil
}

func (s *Storage) ExistsForBrand(ctx context.Context, brandID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	count, err := s.collection.CountDocuments(ctx, activeOnly(bson.M{FieldBrandID: brandID}), options.Count().SetLimit(1))
	if err != nil {
		return false, coreerrs.WrapOperation(err, "check product existence for brand")
	}
	return count > 0, nil
}

func (s *Storage) ExistsForCategory(ctx context.Context, categoryID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	count, err := s.collection.CountDocuments(ctx, activeOnly(bson.M{FieldCategoryID: categoryID}), options.Count().SetLimit(1))
	if err != nil {
		return false, coreerrs.WrapOperation(err, "check product existence for category")
	}
	return count > 0, nil
}

// listSortField picks the sort key for the requested field. Name sorts on
// FieldNameSr since Name itself is a multi-locale map.
func listSortField(sort entities.ProductsListSort) string {
	field := FieldCreatedAt
	if sort.Field == entities.ProductsListSortFieldName {
		field = FieldNameSr
	}
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
