// Package invoice defines the Invoice service interface; the
// implementation lives in the invoice subpackage.
package invoice

import (
	"context"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
)

// Service renders a Sale as an invoice PDF. Stateless — see
// entities.InvoiceGenerate.
type Service interface {
	// Generate returns the rendered PDF's bytes. When in.SendEmail is set,
	// it also emails the PDF to the buyer as a best-effort side effect (a
	// delivery failure is logged, not returned as an error) — see
	// entities.InvoiceGenerate.SendEmail.
	Generate(ctx context.Context, in *entities.InvoiceGenerate) ([]byte, error)
}
