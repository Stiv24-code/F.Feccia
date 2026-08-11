package pdftemplates

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.PdfTemplate{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func baseRequest() dto.PdfTemplateRequest {
	return dto.PdfTemplateRequest{
		Name:    "ACME layout",
		Client:  "ACME S.r.l.",
		Senders: []string{" Ordini@ACME.it ", "@acme.it", ""},
		Fields: []dto.PdfTemplateFieldDTO{
			{Target: "ref", Label: "Riferimento", Page: 0, X: 0.1, Y: 0.1, W: 0.2, H: 0.05},
			{Target: "kg", Label: "Peso", Page: 0, X: 0.5, Y: 0.1, W: 0.2, H: 0.05},
		},
	}
}

func TestPdfTemplateService_CreateNormalizesAndRoundTrips(t *testing.T) {
	ctx := context.Background()
	svc := NewPdfTemplateService(newTestDB(t))

	created, err := svc.Create(ctx, baseRequest())
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	// Senders lowercased/trimmed, empty entries dropped.
	if len(created.Senders) != 2 || created.Senders[0] != "ordini@acme.it" {
		t.Fatalf("expected normalized senders, got %v", created.Senders)
	}
	// Field IDs assigned when missing.
	for _, f := range created.Fields {
		if f.ID == "" {
			t.Fatalf("expected field ID to be assigned, got empty on %+v", f)
		}
	}

	got, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if len(got.Fields) != 2 || got.Fields[0].Target != "ref" || got.Fields[1].W != 0.2 {
		t.Fatalf("expected fields to round-trip through JSON, got %+v", got.Fields)
	}
}

func TestPdfTemplateService_RejectsInvalidInput(t *testing.T) {
	ctx := context.Background()
	svc := NewPdfTemplateService(newTestDB(t))

	// Missing name -> 400.
	req := baseRequest()
	req.Name = "   "
	if _, err := svc.Create(ctx, req); err == nil {
		t.Fatalf("expected error for empty name")
	} else {
		var apiErr utils.APIError
		if !errors.As(err, &apiErr) || apiErr.Code != 400 {
			t.Fatalf("expected APIError 400, got %v", err)
		}
	}

	// Unknown field target -> 400. Targets are snake_case here (they mirror
	// InboundOrder JSON names), so OrderMesh's old camelCase must be refused.
	req = baseRequest()
	req.Fields[0].Target = "loadDate"
	if _, err := svc.Create(ctx, req); err == nil {
		t.Fatalf("expected error for unknown target loadDate")
	}
}

func TestPdfTemplateService_UpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	svc := NewPdfTemplateService(newTestDB(t))

	created, err := svc.Create(ctx, baseRequest())
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	req := baseRequest()
	req.Name = "ACME layout v2"
	req.Fields = req.Fields[:1]
	updated, err := svc.Update(ctx, created.ID, req)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Name != "ACME layout v2" || len(updated.Fields) != 1 {
		t.Fatalf("expected updated template, got %+v", updated)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if err := svc.Delete(ctx, created.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound on second delete, got %v", err)
	}
	if _, err := svc.Update(ctx, uuid.New(), baseRequest()); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound updating missing template, got %v", err)
	}
}

func TestPdfTemplateService_MatchExactBeatsDomain(t *testing.T) {
	ctx := context.Background()
	svc := NewPdfTemplateService(newTestDB(t))

	domainOnly := baseRequest()
	domainOnly.Name = "domain-wide"
	domainOnly.Senders = []string{"@acme.it"}
	if _, err := svc.Create(ctx, domainOnly); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	exact := baseRequest()
	exact.Name = "exact-address"
	exact.Senders = []string{"ordini@acme.it"}
	exactCreated, err := svc.Create(ctx, exact)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	m, err := svc.Match(ctx, "ORDINI@acme.it")
	if err != nil {
		t.Fatalf("Match returned error: %v", err)
	}
	if m == nil || m.ID != exactCreated.ID {
		t.Fatalf("expected exact-address template to win, got %+v", m)
	}

	// Domain match still works for other addresses on the domain.
	m, err = svc.Match(ctx, "altro@acme.it")
	if err != nil {
		t.Fatalf("Match returned error: %v", err)
	}
	if m == nil || m.Name != "domain-wide" {
		t.Fatalf("expected domain template for other address, got %+v", m)
	}

	// No match -> nil, no error.
	m, err = svc.Match(ctx, "chiunque@altrove.it")
	if err != nil {
		t.Fatalf("Match returned error: %v", err)
	}
	if m != nil {
		t.Fatalf("expected nil match, got %+v", m)
	}
}
