// Package salesreport defines the SalesReport service interface; the
// implementation lives in the salesreport subpackage.
package salesreport

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service renders every Sale in a period as an .xlsx workbook. Stateless —
// see entities.SalesReportGenerate.
type Service interface {
	// Generate returns the rendered workbook's bytes.
	Generate(ctx context.Context, in *entities.SalesReportGenerate) ([]byte, error)
}
