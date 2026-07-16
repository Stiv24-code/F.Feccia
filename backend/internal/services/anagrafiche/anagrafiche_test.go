package anagrafiche

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Country{}, &models.Bank{}, &models.AccountingEntry{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestAnagraficheService_CreateCountry_NormalizesIso2(t *testing.T) {
	ctx := context.Background()
	svc := NewAnagraficheService(newTestDB(t))

	resp, err := svc.CreateCountry(ctx, dto.CountryRequest{Iso2: " it ", Nome: "Italia"})
	if err != nil {
		t.Fatalf("CreateCountry returned error: %v", err)
	}
	if resp.Iso2 != "IT" {
		t.Fatalf("expected normalized Iso2 %q, got %q", "IT", resp.Iso2)
	}
}

func TestAnagraficheService_CreateCountry_InvalidIso2(t *testing.T) {
	ctx := context.Background()
	svc := NewAnagraficheService(newTestDB(t))

	_, err := svc.CreateCountry(ctx, dto.CountryRequest{Iso2: "ITA", Nome: "Italia"})
	var apiErr utils.APIError
	if err == nil {
		t.Fatal("expected error for 3-letter iso2")
	}
	if !errorAs(err, &apiErr) || apiErr.Code != 400 {
		t.Fatalf("expected APIError 400, got %v", err)
	}
}

func TestAnagraficheService_CreateCountry_DuplicateReturns409(t *testing.T) {
	ctx := context.Background()
	svc := NewAnagraficheService(newTestDB(t))

	if _, err := svc.CreateCountry(ctx, dto.CountryRequest{Iso2: "IT", Nome: "Italia"}); err != nil {
		t.Fatalf("first CreateCountry returned error: %v", err)
	}

	_, err := svc.CreateCountry(ctx, dto.CountryRequest{Iso2: "it", Nome: "Italia di nuovo"})
	var apiErr utils.APIError
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if !errorAs(err, &apiErr) || apiErr.Code != 409 {
		t.Fatalf("expected APIError 409, got %v", err)
	}
}

func TestAnagraficheService_CreateAccountingEntry_DefaultsTipoAndValidates(t *testing.T) {
	ctx := context.Background()
	svc := NewAnagraficheService(newTestDB(t))

	resp, err := svc.CreateAccountingEntry(ctx, dto.AccountingEntryRequest{Codice: "RIC001", Descrizione: "Trasporto"})
	if err != nil {
		t.Fatalf("CreateAccountingEntry returned error: %v", err)
	}
	if resp.Tipo != "ricavo" {
		t.Fatalf("expected default tipo %q, got %q", "ricavo", resp.Tipo)
	}
	if resp.IvaCodice != "N8" {
		t.Fatalf("expected default iva_codice %q, got %q", "N8", resp.IvaCodice)
	}

	_, err = svc.CreateAccountingEntry(ctx, dto.AccountingEntryRequest{Codice: "X1", Descrizione: "Bad", Tipo: "bogus"})
	var apiErr utils.APIError
	if err == nil || !errorAs(err, &apiErr) || apiErr.Code != 400 {
		t.Fatalf("expected APIError 400 for invalid tipo, got %v", err)
	}
}

func TestAnagraficheService_CreateAccountingEntry_DuplicateCodiceReturns409(t *testing.T) {
	ctx := context.Background()
	svc := NewAnagraficheService(newTestDB(t))

	if _, err := svc.CreateAccountingEntry(ctx, dto.AccountingEntryRequest{Codice: "RIC001", Descrizione: "Trasporto"}); err != nil {
		t.Fatalf("first CreateAccountingEntry returned error: %v", err)
	}

	_, err := svc.CreateAccountingEntry(ctx, dto.AccountingEntryRequest{Codice: "RIC001", Descrizione: "Duplicato"})
	var apiErr utils.APIError
	if err == nil || !errorAs(err, &apiErr) || apiErr.Code != 409 {
		t.Fatalf("expected APIError 409 for duplicate codice, got %v", err)
	}
}

func TestAnagraficheService_Bank_CRUD(t *testing.T) {
	ctx := context.Background()
	svc := NewAnagraficheService(newTestDB(t))

	created, err := svc.CreateBank(ctx, dto.BankRequest{Nome: "Banca Test"})
	if err != nil {
		t.Fatalf("CreateBank returned error: %v", err)
	}

	list, err := svc.ListBanks(ctx, "")
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 bank, got %v (err=%v)", list, err)
	}

	if err := svc.DeleteBank(ctx, created.ID); err != nil {
		t.Fatalf("DeleteBank returned error: %v", err)
	}
	list, err = svc.ListBanks(ctx, "")
	if err != nil || len(list) != 0 {
		t.Fatalf("expected 0 banks after delete, got %v (err=%v)", list, err)
	}
}

func errorAs(err error, target *utils.APIError) bool {
	if apiErr, ok := err.(utils.APIError); ok {
		*target = apiErr
		return true
	}
	return false
}
