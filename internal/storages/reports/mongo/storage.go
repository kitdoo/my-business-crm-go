package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	coreerrs "github.com/altessa-s/go-atlas/core/errors"
	datamongo "github.com/altessa-s/go-atlas/data/mongo"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/storages/reports"
)

// Collection and field names here mirror storages/sales/mongo and
// storages/partners/mongo. There is no shared schema package for
// aggregation pipelines to import against, so these are intentionally
// duplicated; keep them in sync if those schemas change.
const (
	salesCollectionName    = "sales"
	partnersCollectionName = "partners"

	fieldCreatedAt           = "created_at"
	fieldCreatedBy           = "created_by"
	fieldStatus              = "status"
	fieldPartnerID           = "partner_id"
	fieldTotalAmount         = "total_amount"
	fieldItems               = "items"
	fieldItemVariantID       = "items.variant_id"
	fieldItemQuantity        = "items.quantity"
	fieldItemPriceAmount     = "items.price_amount"
	fieldItemDiscountPercent = "items.discount_percentage"

	fieldPartnerCommissionPercentage = "commission_percentage"
)

var _ reports.Storage = (*Storage)(nil)

// Storage implements reports.Storage as MongoDB aggregation pipelines.
type Storage struct {
	sales    *mongo.Collection
	partners *mongo.Collection
}

// New builds a Storage backed by db's "sales" and "partners" collections.
func New(db *datamongo.Mongo) *Storage {
	return &Storage{
		sales:    db.Database().Collection(salesCollectionName),
		partners: db.Database().Collection(partnersCollectionName),
	}
}

// periodMatch builds the shared match stage: createdAt within period and
// status not cancelled — a cancelled sale shouldn't count toward revenue
// reports (the TD doesn't state this explicitly; it's the one judgment
// call these pipelines make).
func periodMatch(period *entities.PeriodFilter) bson.D {
	rng := bson.M{}
	if period != nil {
		if period.From != nil {
			rng["$gte"] = *period.From
		}
		if period.To != nil {
			rng["$lte"] = *period.To
		}
	}
	filter := bson.M{fieldStatus: bson.M{"$ne": int32(entities.SaleStatusCancelled)}}
	if len(rng) > 0 {
		filter[fieldCreatedAt] = rng
	}
	return bson.D{{Key: "$match", Value: filter}}
}

func (s *Storage) GetSalesReport(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesReportRow, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	pipeline := mongo.Pipeline{
		periodMatch(period),
		{{Key: "$group", Value: bson.M{
			"_id":         bson.M{"$dateTrunc": bson.M{"date": "$" + fieldCreatedAt, "unit": "day"}},
			"salesCount":  bson.M{"$sum": 1},
			"totalAmount": bson.M{"$sum": "$" + fieldTotalAmount},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	var out []struct {
		PeriodStart time.Time `bson:"_id"`
		SalesCount  int64     `bson:"salesCount"`
		TotalAmount int64     `bson:"totalAmount"`
	}
	if err := runAggregate(ctx, s.sales, pipeline, &out); err != nil {
		return nil, coreerrs.WrapOperation(err, "aggregate sales report")
	}

	rows := make([]entities.SalesReportRow, 0, len(out))
	for _, r := range out {
		rows = append(rows, entities.SalesReportRow{
			PeriodStart: r.PeriodStart,
			PeriodEnd:   r.PeriodStart.Add(24 * time.Hour),
			SalesCount:  r.SalesCount,
			TotalAmount: r.TotalAmount,
		})
	}
	return rows, nil
}

func (s *Storage) GetSalesByStaff(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesByStaffRow, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	pipeline := mongo.Pipeline{
		periodMatch(period),
		{{Key: "$group", Value: bson.M{
			"_id":         "$" + fieldCreatedBy,
			"salesCount":  bson.M{"$sum": 1},
			"totalAmount": bson.M{"$sum": "$" + fieldTotalAmount},
		}}},
		{{Key: "$sort", Value: bson.M{"totalAmount": -1}}},
	}

	var out []struct {
		UserID      string `bson:"_id"`
		SalesCount  int64  `bson:"salesCount"`
		TotalAmount int64  `bson:"totalAmount"`
	}
	if err := runAggregate(ctx, s.sales, pipeline, &out); err != nil {
		return nil, coreerrs.WrapOperation(err, "aggregate sales by staff")
	}

	rows := make([]entities.SalesByStaffRow, 0, len(out))
	for _, r := range out {
		rows = append(rows, entities.SalesByStaffRow{UserID: r.UserID, SalesCount: r.SalesCount, TotalAmount: r.TotalAmount})
	}
	return rows, nil
}

func (s *Storage) GetSalesByPartner(ctx context.Context, period *entities.PeriodFilter) ([]entities.SalesByPartnerRow, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	match := periodMatch(period)
	match[0].Value.(bson.M)[fieldPartnerID] = bson.M{"$ne": nil}

	pipeline := mongo.Pipeline{
		match,
		{{Key: "$group", Value: bson.M{
			"_id":         "$" + fieldPartnerID,
			"salesCount":  bson.M{"$sum": 1},
			"totalAmount": bson.M{"$sum": "$" + fieldTotalAmount},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         partnersCollectionName,
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "partner",
		}}},
		{{Key: "$unwind", Value: "$partner"}},
		{{Key: "$addFields", Value: bson.M{
			"commissionAmount": bson.M{"$toLong": bson.M{"$divide": bson.A{
				bson.M{"$multiply": bson.A{"$totalAmount", "$partner." + fieldPartnerCommissionPercentage}},
				100,
			}}},
		}}},
		{{Key: "$sort", Value: bson.M{"totalAmount": -1}}},
	}

	var out []struct {
		PartnerID        string `bson:"_id"`
		SalesCount       int64  `bson:"salesCount"`
		TotalAmount      int64  `bson:"totalAmount"`
		CommissionAmount int64  `bson:"commissionAmount"`
	}
	if err := runAggregate(ctx, s.sales, pipeline, &out); err != nil {
		return nil, coreerrs.WrapOperation(err, "aggregate sales by partner")
	}

	rows := make([]entities.SalesByPartnerRow, 0, len(out))
	for _, r := range out {
		rows = append(rows, entities.SalesByPartnerRow{
			PartnerID:        r.PartnerID,
			SalesCount:       r.SalesCount,
			TotalAmount:      r.TotalAmount,
			CommissionAmount: r.CommissionAmount,
		})
	}
	return rows, nil
}

func (s *Storage) GetPopularProducts(ctx context.Context, period *entities.PeriodFilter, limit int32) ([]entities.PopularProductRow, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	pipeline := mongo.Pipeline{
		periodMatch(period),
		{{Key: "$unwind", Value: "$" + fieldItems}},
		{{Key: "$group", Value: bson.M{
			"_id":          "$" + fieldItemVariantID,
			"quantitySold": bson.M{"$sum": "$" + fieldItemQuantity},
			// $divide always returns a double in MongoDB's aggregation
			// pipeline, even when both operands are integers — decoding
			// that into the int64 TotalAmount field below fails outright
			// whenever the division isn't exact (driver refuses the
			// lossy float64->int64 truncation). $trunc dividing before
			// summing matches the truncating integer division the sale
			// service itself uses to compute a line total
			// (internal/services/sale/sale/service.go), then $toLong
			// converts the now-integral double into a real BSON Long so
			// $sum accumulates as Long throughout instead of drifting
			// back to double.
			"totalAmount": bson.M{"$sum": bson.M{"$toLong": bson.M{"$trunc": bson.M{"$divide": bson.A{
				bson.M{"$multiply": bson.A{
					"$" + fieldItemPriceAmount,
					"$" + fieldItemQuantity,
					bson.M{"$subtract": bson.A{100, "$" + fieldItemDiscountPercent}},
				}},
				100,
			}}}}},
		}}},
		{{Key: "$sort", Value: bson.M{"quantitySold": -1}}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	var out []struct {
		VariantID    string `bson:"_id"`
		QuantitySold int64  `bson:"quantitySold"`
		TotalAmount  int64  `bson:"totalAmount"`
	}
	if err := runAggregate(ctx, s.sales, pipeline, &out); err != nil {
		return nil, coreerrs.WrapOperation(err, "aggregate popular products")
	}

	rows := make([]entities.PopularProductRow, 0, len(out))
	for _, r := range out {
		rows = append(rows, entities.PopularProductRow{VariantID: r.VariantID, QuantitySold: r.QuantitySold, TotalAmount: r.TotalAmount})
	}
	return rows, nil
}

func (s *Storage) GetTurnover(ctx context.Context, period *entities.PeriodFilter) ([]entities.TurnoverRow, error) {
	ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
	defer cancel()

	pipeline := mongo.Pipeline{
		periodMatch(period),
		{{Key: "$group", Value: bson.M{
			"_id":         bson.M{"$dateTrunc": bson.M{"date": "$" + fieldCreatedAt, "unit": "day"}},
			"totalAmount": bson.M{"$sum": "$" + fieldTotalAmount},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	var out []struct {
		PeriodStart time.Time `bson:"_id"`
		TotalAmount int64     `bson:"totalAmount"`
	}
	if err := runAggregate(ctx, s.sales, pipeline, &out); err != nil {
		return nil, coreerrs.WrapOperation(err, "aggregate turnover")
	}

	rows := make([]entities.TurnoverRow, 0, len(out))
	for _, r := range out {
		rows = append(rows, entities.TurnoverRow{
			PeriodStart: r.PeriodStart,
			PeriodEnd:   r.PeriodStart.Add(24 * time.Hour),
			TotalAmount: r.TotalAmount,
		})
	}
	return rows, nil
}

func runAggregate(ctx context.Context, coll *mongo.Collection, pipeline mongo.Pipeline, out any) error {
	cur, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	return cur.All(ctx, out)
}
