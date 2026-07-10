package report

import (
	"context"

	"google.golang.org/grpc"

	"github.com/altessa-s/go-atlas/domain/converter"
	"github.com/altessa-s/go-atlas/domain/converter/codec/unixtime"

	coreslices "github.com/altessa-s/go-atlas/core/collections/slices"

	commonpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/common"
	reportpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/types/report"

	reportsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/report/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	reportsvc "github.com/kitdoo/my-business-crm-go/internal/services/report"
)

// withUnixTimeCodec bridges int64 Unix-second proto fields (periodStart,
// periodEnd) with time.Time entity fields — see
// PROTO_DEVELOPMENT_STANDARD.md § Timestamps.
var withUnixTimeCodec = converter.WithCodecs(unixtime.New())

// Handler implements reportsvcpb.ReportsServiceServer.
type Handler struct {
	reportsvcpb.UnimplementedReportsServiceServer
	svc reportsvc.Service
}

// New builds a Handler.
func New(svc reportsvc.Service) *Handler { return &Handler{svc: svc} }

// Register attaches the handler to gs.
func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
	reportsvcpb.RegisterReportsServiceServer(gs, h)
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toPeriodFilter(p *commonpb.PeriodFilter) *entities.PeriodFilter {
	if p == nil {
		return nil
	}
	return converter.Convert(p, &entities.PeriodFilter{}, withUnixTimeCodec)
}

func (h *Handler) GetSalesReport(ctx context.Context, in *reportsvcpb.GetSalesReportRequest) (*reportsvcpb.GetSalesReportResponse, error) {
	rows, err := h.svc.GetSalesReport(ctx, toPeriodFilter(in.GetPeriod()))
	if err != nil {
		return nil, err
	}
	return &reportsvcpb.GetSalesReportResponse{
		Rows: coreslices.To(rows, func(r entities.SalesReportRow) *reportpb.SalesReportRow {
			return converter.Convert(&r, &reportpb.SalesReportRow{}, withUnixTimeCodec)
		}),
	}, nil
}

func (h *Handler) GetSalesByStaff(ctx context.Context, in *reportsvcpb.GetSalesByStaffRequest) (*reportsvcpb.GetSalesByStaffResponse, error) {
	rows, err := h.svc.GetSalesByStaff(ctx, toPeriodFilter(in.GetPeriod()))
	if err != nil {
		return nil, err
	}
	return &reportsvcpb.GetSalesByStaffResponse{
		Rows: coreslices.To(rows, func(r entities.SalesByStaffRow) *reportpb.SalesByStaffRow {
			return converter.Convert(&r, &reportpb.SalesByStaffRow{})
		}),
	}, nil
}

func (h *Handler) GetSalesByPartner(ctx context.Context, in *reportsvcpb.GetSalesByPartnerRequest) (*reportsvcpb.GetSalesByPartnerResponse, error) {
	rows, err := h.svc.GetSalesByPartner(ctx, toPeriodFilter(in.GetPeriod()))
	if err != nil {
		return nil, err
	}
	return &reportsvcpb.GetSalesByPartnerResponse{
		Rows: coreslices.To(rows, func(r entities.SalesByPartnerRow) *reportpb.SalesByPartnerRow {
			return converter.Convert(&r, &reportpb.SalesByPartnerRow{})
		}),
	}, nil
}

func (h *Handler) GetPopularProducts(ctx context.Context, in *reportsvcpb.GetPopularProductsRequest) (*reportsvcpb.GetPopularProductsResponse, error) {
	rows, err := h.svc.GetPopularProducts(ctx, toPeriodFilter(in.GetPeriod()), in.GetLimit())
	if err != nil {
		return nil, err
	}
	return &reportsvcpb.GetPopularProductsResponse{
		Rows: coreslices.To(rows, func(r entities.PopularProductRow) *reportpb.PopularProductRow {
			return converter.Convert(&r, &reportpb.PopularProductRow{})
		}),
	}, nil
}

func (h *Handler) GetTurnover(ctx context.Context, in *reportsvcpb.GetTurnoverRequest) (*reportsvcpb.GetTurnoverResponse, error) {
	rows, err := h.svc.GetTurnover(ctx, toPeriodFilter(in.GetPeriod()))
	if err != nil {
		return nil, err
	}
	return &reportsvcpb.GetTurnoverResponse{
		Rows: coreslices.To(rows, func(r entities.TurnoverRow) *reportpb.TurnoverRow {
			return converter.Convert(&r, &reportpb.TurnoverRow{}, withUnixTimeCodec)
		}),
	}, nil
}

func (h *Handler) GetStockLevels(ctx context.Context, in *reportsvcpb.GetStockLevelsRequest) (*reportsvcpb.GetStockLevelsResponse, error) {
	rows, err := h.svc.GetStockLevels(ctx, optionalString(in.GetWarehouseId()))
	if err != nil {
		return nil, err
	}
	return &reportsvcpb.GetStockLevelsResponse{
		Rows: coreslices.To(rows, func(r entities.StockLevelRow) *reportpb.StockLevelRow {
			return converter.Convert(&r, &reportpb.StockLevelRow{})
		}),
	}, nil
}
