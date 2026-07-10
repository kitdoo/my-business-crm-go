// Package category implements the category.Service interface.
package category

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/altessa-s/go-atlas/domain/normalizer"
	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	categorysvc "github.com/kitdoo/my-business-crm-go/internal/services/category"
	"github.com/kitdoo/my-business-crm-go/internal/storages/categories"
)

var _ categorysvc.Service = (*Service)(nil)

// Service is the category.Service implementation.
type Service struct {
	storage  categories.Storage
	products categorysvc.ProductsExistenceChecker // optional; nil until the products aggregate exists
	logger   *slog.Logger
}

// New builds a Service. products may be nil (see ProductsExistenceChecker).
func New(storage categories.Storage, products categorysvc.ProductsExistenceChecker) *Service {
	return &Service{
		storage:  storage,
		products: products,
		logger:   slog.Default().With(slogx.Module("service:category")),
	}
}

// validateParent rejects a self-referencing parent and confirms parentID
// resolves to an existing category, translating a not-found into
// errs.ErrCategoryParentNotFound so it is distinguishable from a missing
// subject category.
func (s *Service) validateParent(ctx context.Context, id string, parentID *string) error {
	if parentID == nil {
		return nil
	}
	if *parentID == id {
		return errs.ErrCategorySelfParent
	}
	if _, err := s.storage.Get(ctx, *parentID); err != nil {
		if errors.Is(err, errs.ErrCategoryNotFound) {
			return errs.ErrCategoryParentNotFound
		}
		return err
	}
	return nil
}

func (s *Service) Create(ctx context.Context, in *entities.CategoryCreate) (*entities.Category, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	c := entities.CategoryNew()
	if err := s.validateParent(ctx, c.ID, in.ParentID); err != nil {
		return nil, err
	}
	in.Merge(c)

	if err := s.storage.Insert(ctx, c); err != nil {
		s.logger.DebugContext(ctx, "insert category failed", slogx.Error(err))
		return nil, err
	}
	return c, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entities.Category, error) {
	return s.storage.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, in *entities.CategoriesList) (*entities.List[entities.Category], error) {
	return s.storage.List(ctx, in)
}

func (s *Service) Update(ctx context.Context, in *entities.CategoryUpdate) (*entities.Category, error) {
	_ = normalizer.Normalize(in) //nolint:errcheck

	c, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Etag != nil && *in.Etag != c.Etag {
		return nil, errs.ErrStaleEntity
	}
	if err := s.validateParent(ctx, c.ID, in.ParentID); err != nil {
		return nil, err
	}
	oldEtag := c.Etag
	in.Merge(c)
	c.BeforeUpdate()
	if err := s.storage.Update(ctx, c, oldEtag); err != nil {
		s.logger.DebugContext(ctx, "update category failed", slog.String("id", c.ID), slogx.Error(err))
		return nil, err
	}
	return c, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.CategoryDelete) error {
	_ = normalizer.Normalize(in) //nolint:errcheck

	c, err := s.storage.Get(ctx, in.ID)
	if err != nil {
		return err
	}
	if in.Etag != nil && *in.Etag != c.Etag {
		return errs.ErrStaleEntity
	}

	if s.products != nil {
		hasProducts, err := s.products.ExistsForCategory(ctx, c.ID)
		if err != nil {
			return err
		}
		if hasProducts {
			return errs.ErrCategoryHasProducts
		}
	}

	oldEtag := c.Etag
	now := time.Now().UTC()
	c.DeletedAt = &now
	c.BeforeUpdate()
	if err := s.storage.SoftDelete(ctx, &entities.SoftDelete{
		ID:           c.ID,
		Etag:         oldEtag,
		NewUpdatedAt: *c.DeletedAt,
		NewEtag:      c.Etag,
	}); err != nil {
		s.logger.DebugContext(ctx, "soft delete category failed", slog.String("id", c.ID), slogx.Error(err))
		return err
	}
	return nil
}
