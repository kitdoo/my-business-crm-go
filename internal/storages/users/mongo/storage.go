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
	"github.com/kitdoo/my-business-crm-go/internal/storages/users"
)

const collectionName = "users"

const (
	FieldID = "_id"
	// FieldNameSr is the sort key for Name. "sr" is the mandatory locale
	// (see common_localized_string.proto), used as a deterministic single
	// key since Name itself is a multi-locale map.
	FieldNameSr    = "name.sr"
	FieldPhone     = "phone"
	FieldEmail     = "email"
	FieldRole      = "role"
	FieldStatus    = "status"
	FieldCreatedAt = "created_at"
	FieldUpdatedAt = "updated_at"
	FieldDeletedAt = "deleted_at"
	FieldEtag      = "etag"
	FieldCursorId  = "cursor_id"
)

const defaultListLimit = datamongo.DefaultListLimit

type model struct {
	ID           string                   `bson:"_id"`
	Name         entities.LocalizedString `bson:"name"`
	LastName     entities.LocalizedString `bson:"last_name"`
	Phone        string                   `bson:"phone"`
	Email        string                   `bson:"email"`
	Description  entities.LocalizedString `bson:"description"`
	Role         entities.UserRole        `bson:"role"`
	Status       entities.UserStatus      `bson:"status"`
	PasswordHash string                   `bson:"password_hash"`
	CreatedAt    time.Time                `bson:"created_at,omitonupdate"`
	UpdatedAt    time.Time                `bson:"updated_at"`
	DeletedAt    *time.Time               `bson:"deleted_at,omitonupdate"`
	Etag         string                   `bson:"etag"`
	CursorId     bson.ObjectID            `bson:"cursor_id,omitonupdate"`
}

var _ users.Storage = (*Storage)(nil)

// Storage implements users.Storage against MongoDB.
type Storage struct {
	db         *datamongo.Mongo
	collection *mongo.Collection
}

// New builds a Storage backed by db's "users" collection.
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
	switch {
	case fields.Contains(FieldPhone):
		return errs.ErrUserPhoneConflict
	case fields.Contains(FieldEmail):
		return errs.ErrUserEmailConflict
	default:
		return nil
	}
}

func (s *Storage) Insert(ctx context.Context, u *entities.User) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	m := converter.Convert(u, &model{}, converter.WithHandleEmbeddedStructs(true))
	m.CursorId = bson.NewObjectID()
	doc, err := s.db.ConvertToNewDocument(ctx, m)
	if err != nil {
		return coreerrs.WrapOperation(err, "convert user document")
	}
	if _, err := s.collection.InsertOne(ctx, doc); err != nil {
		if dupErr := classifyDuplicate(err); dupErr != nil {
			return dupErr
		}
		return coreerrs.WrapOperation(err, "insert user")
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, id string) (*entities.User, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	var m model
	if err := s.collection.FindOne(ctx, activeOnly(bson.M{FieldID: id})).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrUserNotFound
		}
		return nil, coreerrs.WrapOperation(err, "get user")
	}
	return converter.Convert(&m, &entities.User{}), nil
}

func (s *Storage) GetByLogin(ctx context.Context, login string) (*entities.User, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	filter := activeOnly(bson.M{"$or": bson.A{
		bson.M{FieldPhone: login},
		bson.M{FieldEmail: login},
	}})
	var m model
	if err := s.collection.FindOne(ctx, filter).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrUserNotFound
		}
		return nil, coreerrs.WrapOperation(err, "get user by login")
	}
	return converter.Convert(&m, &entities.User{}), nil
}

func (s *Storage) Update(ctx context.Context, u *entities.User, oldEtag string) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	update, err := s.db.ConvertToUpdateDocument(ctx, converter.Convert(u, &model{}))
	if err != nil {
		return coreerrs.WrapOperation(err, "build update document")
	}
	filter := activeOnly(bson.M{FieldID: u.ID})
	if oldEtag != "" {
		filter[FieldEtag] = oldEtag
	}
	res, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		if dupErr := classifyDuplicate(err); dupErr != nil {
			return dupErr
		}
		return coreerrs.WrapOperation(err, "update user")
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
		return coreerrs.WrapOperation(err, "soft delete user")
	}
	if res.MatchedCount == 0 {
		if in.Etag != "" {
			return errs.ErrStaleEntity
		}
		return errs.ErrUserNotFound
	}
	return nil
}

func (s *Storage) List(ctx context.Context, in *entities.UsersList) (*entities.List[entities.User], error) {
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
	if len(in.Roles) > 0 {
		filter[FieldRole] = bson.M{"$in": in.Roles}
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
		return nil, coreerrs.WrapOperation(err, "list users")
	}

	out := &entities.List[entities.User]{
		Items: coreslices.To(result.Items, func(m model) *entities.User {
			return converter.Convert(&m, &entities.User{})
		}),
		Total: result.Total,
	}
	if result.NextCursor != nil {
		out.NextCursor = *result.NextCursor
	}
	return out, nil
}

// listSortField picks the sort key for the requested field. Name sorts on
// FieldNameSr since Name itself is a multi-locale map.
func listSortField(sort entities.UsersListSort) string {
	field := FieldCreatedAt
	if sort.Field == entities.UsersListSortFieldName {
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
