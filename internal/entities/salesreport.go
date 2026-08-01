package entities

import "time"

// SalesReportGenerate is SalesReportService.Generate's input — renders
// every Sale created within [From, To] (both inclusive) as an .xlsx
// workbook: one row per sale plus a totals row. Stateless, like
// InvoiceGenerate: nothing is persisted, the workbook is rendered fresh
// from the current Sale/Client/Partner/User data every time.
type SalesReportGenerate struct {
	From time.Time
	To   time.Time
}
