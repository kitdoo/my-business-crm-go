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
	"github.com/kitdoo/my-business-crm-go/internal/storages/productattributedefinitions"
)

const collectionName = "product_attribute_definitions"

const (
	FieldID        = "_id"
	FieldKey       = "key"
	FieldLabel     = "label"
	FieldIsPublic  = "is_public"
	FieldSortOrder = "sort_order"
	FieldCreatedAt = "created_at"
	FieldCursorId  = "cursor_id"
)

const defaultListLimit = datamongo.DefaultListLimit

type model struct {
	ID        string                   `bson:"_id"`
	Key       string                   `bson:"key"`
	Label     entities.LocalizedString `bson:"label"`
	IsPublic  bool                     `bson:"is_public"`
	SortOrder int32                    `bson:"sort_order"`
	CreatedAt time.Time                `bson:"created_at"`
	CursorId  bson.ObjectID            `bson:"cursor_id"`
}

var _ productattributedefinitions.Storage = (*Storage)(nil)

// Storage implements productattributedefinitions.Storage against MongoDB.
type Storage struct {
	db         *datamongo.Mongo
	collection *mongo.Collection
}

// New builds a Storage backed by db's collection.
func New(db *datamongo.Mongo) *Storage {
	return &Storage{db: db, collection: db.Database().Collection(collectionName)}
}

func (s *Storage) Get(ctx context.Context, id string) (*entities.ProductAttributeDefinition, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	var m model
	err := s.collection.FindOne(ctx, bson.M{FieldID: id}).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errs.ErrProductAttributeDefinitionNotFound
	}
	if err != nil {
		return nil, coreerrs.WrapOperation(err, "get product attribute definition")
	}
	return converter.Convert(&m, &entities.ProductAttributeDefinition{}), nil
}

func (s *Storage) List(ctx context.Context, in *entities.ProductAttributeDefinitionsList) (*entities.List[entities.ProductAttributeDefinition], error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	limit := in.Pagination.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	filter := bson.M{}
	if in.IsPublic != nil {
		filter[FieldIsPublic] = *in.IsPublic
	}

	opts := []datamongo.ListCursorOption{
		datamongo.WithListCursorFilter(filter),
		datamongo.WithListCursorSort(FieldSortOrder),
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
		return nil, coreerrs.WrapOperation(err, "list product attribute definitions")
	}

	out := &entities.List[entities.ProductAttributeDefinition]{
		Items: coreslices.To(result.Items, func(m model) *entities.ProductAttributeDefinition {
			return converter.Convert(&m, &entities.ProductAttributeDefinition{})
		}),
	}
	if result.NextCursor != nil {
		out.NextCursor = *result.NextCursor
	}
	return out, nil
}
