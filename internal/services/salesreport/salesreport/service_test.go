package salesreport

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/pkg/appconfig"
	clientsvc "github.com/kitdoo/my-business-crm-go/internal/services/client"
	partnersvc "github.com/kitdoo/my-business-crm-go/internal/services/partner"
	salesvc "github.com/kitdoo/my-business-crm-go/internal/services/sale"
	usersvc "github.com/kitdoo/my-business-crm-go/internal/services/user"
)

// fakeSales/fakeClients/fakePartners/fakeUsers are minimal stand-ins for
// their respective Service interfaces — TestGenerate only exercises List
// (fakeSales) and Get (the other three), so every other method just
// panics if ever called.
type fakeSales struct {
	salesvc.Service
	items []*entities.Sale
}

func (f *fakeSales) List(_ context.Context, _ *entities.SalesList) (*entities.List[entities.Sale], error) {
	return &entities.List[entities.Sale]{Items: f.items}, nil
}

type fakeClients struct {
	clientsvc.Service
	byID map[string]*entities.Client
}

func (f *fakeClients) Get(_ context.Context, id string) (*entities.Client, error) {
	return f.byID[id], nil
}

type fakePartners struct {
	partnersvc.Service
	byID map[string]*entities.Partner
}

func (f *fakePartners) Get(_ context.Context, id string) (*entities.Partner, error) {
	return f.byID[id], nil
}

type fakeUsers struct {
	usersvc.Service
	byID map[string]*entities.User
}

func (f *fakeUsers) Get(_ context.Context, id string) (*entities.User, error) {
	return f.byID[id], nil
}

func TestGenerate(t *testing.T) {
	partnerID := "partner-1"
	sales := &fakeSales{items: []*entities.Sale{
		{
			Number: 1, ClientID: "client-1", TotalAmount: 150000, Status: entities.SaleStatusCompleted,
			CreatedBy: "user-1", CreatedAt: time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			Number: 2, PartnerID: &partnerID, TotalAmount: 50000, Status: entities.SaleStatusDraft,
			CreatedBy: "user-1", CreatedAt: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		},
	}}
	clients := &fakeClients{byID: map[string]*entities.Client{
		"client-1": {ID: "client-1", Name: "Marko Marković"},
	}}
	partners := &fakePartners{byID: map[string]*entities.Partner{
		partnerID: {ID: partnerID, Name: "ABC Gradnja doo"},
	}}
	users := &fakeUsers{byID: map[string]*entities.User{
		"user-1": {ID: "user-1", Name: entities.LocalizedString{"sr": "Jovan"}, LastName: entities.LocalizedString{"sr": "Tomić"}},
	}}

	s := New(sales, clients, partners, users, &appconfig.Config{CRM: &appconfig.CRMConfig{Currency: "RSD", DefaultLocale: "sr"}})

	data, err := s.Generate(context.Background(), &entities.SalesReportGenerate{
		From: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty workbook")
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("reopen generated workbook: %v", err)
	}
	defer f.Close() //nolint:errcheck

	rows, err := f.GetRows("Sales")
	if err != nil {
		t.Fatalf("read rows: %v", err)
	}
	// header + 2 sale rows + 1 totals row.
	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d: %v", len(rows), rows)
	}
	if rows[1][2] != "Marko Marković" {
		t.Errorf("row 1 buyer = %q, want client name", rows[1][2])
	}
	if rows[1][3] != "Jovan Tomić" {
		t.Errorf("row 1 seller = %q, want resolved user name", rows[1][3])
	}
	if rows[2][2] != "ABC Gradnja doo" {
		t.Errorf("row 2 buyer = %q, want partner name", rows[2][2])
	}
	if rows[3][4] != "2 prodaja" {
		t.Errorf("totals row sales count = %q", rows[3][4])
	}
}
