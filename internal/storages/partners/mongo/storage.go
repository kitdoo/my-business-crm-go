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
	"github.com/kitdoo/my-business-crm-go/internal/storages/partners"
)

const collectionName = "partners"

const (
	FieldID        = "_id"
	FieldName      = "name"
	FieldPhone     = "phone"
	FieldStatus    = "status"
	FieldCreatedAt = "created_at"
	FieldUpdatedAt = "updated_at"
	FieldDeletedAt = "deleted_at"
	FieldEtag      = "etag"
	FieldCursorId  = "cursor_id"
)

const defaultListLimit = datamongo.DefaultListLimit

type model struct {
	ID                   string                 `bson:"_id"`
	Name                 string                 `bson:"name"`
	Phone                string                 `bson:"phone"`
	Email                string                 `bson:"email"`
	CommissionPercentage int32                  `bson:"commission_percentage"`
	Note                 string                 `bson:"note"`
	Status               entities.PartnerStatus `bson:"status"`
	CreatedAt            time.Time              `bson:"created_at,omitonupdate"`
	UpdatedAt            time.Time              `bson:"updated_at"`
	DeletedAt            *time.Time             `bson:"deleted_at,omitonupdate"`
	Etag                 string                 `bson:"etag"`
	CursorId             bson.ObjectID          `bson:"cursor_id,omitonupdate"`
}

var _ partners.Storage = (*Storage)(nil)

// Storage implements partners.Storage against MongoDB.
type Storage struct {
	db         *datamongo.Mongo
	collection *mongo.Collection
}

// New builds a Storage backed by db's "partners" collection.
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

// classifyDuplicate maps a Mongo duplicate-key error to a domain sentinel by
// the conflicting field, never by string matching.
func classifyDuplicate(err error) error {
	isDup, fields := datamongo.IsErrorDuplicate(err)
	if !isDup || fields == nil {
		return nil
	}
	if fields.Contains(FieldPhone) {
		return errs.ErrPartnerPhoneConflict
	}
	return nil
}

func (s *Storage) Insert(ctx context.Context, p *entities.Partner) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	m := converter.Convert(p, &model{}, converter.WithHandleEmbeddedStructs(true))
	m.CursorId = bson.NewObjectID()
	doc, err := s.db.ConvertToNewDocument(ctx, m)
	if err != nil {
		return coreerrs.WrapOperation(err, "convert partner document")
	}
	if _, err := s.collection.InsertOne(ctx, doc); err != nil {
		if dupErr := classifyDuplicate(err); dupErr != nil {
			return dupErr
		}
		return coreerrs.WrapOperation(err, "insert partner")
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, id string) (*entities.Partner, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	var m model
	if err := s.collection.FindOne(ctx, activeOnly(bson.M{FieldID: id})).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrPartnerNotFound
		}
		return nil, coreerrs.WrapOperation(err, "get partner")
	}
	return converter.Convert(&m, &entities.Partner{}), nil
}

func (s *Storage) Update(ctx context.Context, p *entities.Partner, oldEtag string) error {
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
		if dupErr := classifyDuplicate(err); dupErr != nil {
			return dupErr
		}
		return coreerrs.WrapOperation(err, "update partner")
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
		return coreerrs.WrapOperation(err, "soft delete partner")
	}
	if res.MatchedCount == 0 {
		if in.Etag != "" {
			return errs.ErrStaleEntity
		}
		return errs.ErrPartnerNotFound
	}
	return nil
}

func (s *Storage) List(ctx context.Context, in *entities.PartnersList) (*entities.List[entities.Partner], error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	limit := in.Pagination.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	filter := activeOnly(bson.M{})
	if len(in.Statuses) > 0 {
		filter[FieldStatus] = bson.M{"$in": in.Statuses}
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
		return nil, coreerrs.WrapOperation(err, "list partners")
	}

	out := &entities.List[entities.Partner]{
		Items: coreslices.To(result.Items, func(m model) *entities.Partner {
			return converter.Convert(&m, &entities.Partner{})
		}),
		Total: result.Total,
	}
	if result.NextCursor != nil {
		out.NextCursor = *result.NextCursor
	}
	return out, nil
}

// listSortField picks the sort key for the requested field.
func listSortField(sort entities.PartnersListSort) string {
	field := FieldCreatedAt
	if sort.Field == entities.PartnersListSortFieldName {
		field = FieldName
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
