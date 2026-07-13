package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/altessa-s/go-atlas/domain/converter"

	coreerrs "github.com/altessa-s/go-atlas/core/errors"
	datamongo "github.com/altessa-s/go-atlas/data/mongo"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	"github.com/kitdoo/my-business-crm-go/internal/storages/images"
)

const collectionName = "images"

const (
	FieldID          = "_id"
	FieldContentType = "content_type"
	FieldSizeBytes   = "size_bytes"
	FieldCreatedAt   = "created_at"
)

type model struct {
	ID          string    `bson:"_id"`
	ContentType string    `bson:"content_type"`
	SizeBytes   int64     `bson:"size_bytes"`
	CreatedAt   time.Time `bson:"created_at,omitonupdate"`
}

var _ images.Storage = (*Storage)(nil)

// Storage implements images.Storage against MongoDB.
type Storage struct {
	db         *datamongo.Mongo
	collection *mongo.Collection
}

// New builds a Storage backed by db's "images" collection.
func New(db *datamongo.Mongo) *Storage {
	return &Storage{
		db:         db,
		collection: db.Database().Collection(collectionName),
	}
}

func (s *Storage) Insert(ctx context.Context, img *entities.Image) error {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	m := converter.Convert(img, &model{})
	doc, err := s.db.ConvertToNewDocument(ctx, m)
	if err != nil {
		return coreerrs.WrapOperation(err, "convert image document")
	}
	if _, err := s.collection.InsertOne(ctx, doc); err != nil {
		return coreerrs.WrapOperation(err, "insert image")
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, id string) (*entities.Image, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	var m model
	if err := s.collection.FindOne(ctx, bson.M{FieldID: id}).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrImageNotFound
		}
		return nil, coreerrs.WrapOperation(err, "get image")
	}
	return converter.Convert(&m, &entities.Image{}), nil
}
