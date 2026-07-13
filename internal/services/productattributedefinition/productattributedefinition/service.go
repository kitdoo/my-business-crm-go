// Package productattributedefinition implements the
// productattributedefinition.Service interface.
package productattributedefinition

import (
	"context"
	"log/slog"

	slogx "github.com/altessa-s/go-atlas/observability/slog"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	productattributedefinitionsvc "github.com/kitdoo/my-business-crm-go/internal/services/productattributedefinition"
	"github.com/kitdoo/my-business-crm-go/internal/storages/productattributedefinitions"
)

var _ productattributedefinitionsvc.Service = (*Service)(nil)

// Service is the productattributedefinition.Service implementation.
type Service struct {
	storage productattributedefinitions.Storage
	logger  *slog.Logger
}

// New builds a Service.
func New(storage productattributedefinitions.Storage) *Service {
	return &Service{
		storage: storage,
		logger:  slog.Default().With(slogx.Module("service:productattributedefinition")),
	}
}

func (s *Service) Get(ctx context.Context, id string) (*entities.ProductAttributeDefinition, error) {
	d, err := s.storage.Get(ctx, id)
	if err != nil {
		s.logger.DebugContext(ctx, "get product attribute definition failed", slog.String("id", id), slogx.Error(err))
		return nil, err
	}
	return d, nil
}

func (s *Service) List(ctx context.Context, in *entities.ProductAttributeDefinitionsList) (*entities.List[entities.ProductAttributeDefinition], error) {
	list, err := s.storage.List(ctx, in)
	if err != nil {
		s.logger.DebugContext(ctx, "list product attribute definitions failed", slogx.Error(err))
		return nil, err
	}
	return list, nil
}
