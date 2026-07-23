package invoices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/database"
	"fratelli-feccia/pkg/pdfgen"
	"fratelli-feccia/pkg/s3invoices"
	"fratelli-feccia/pkg/utils"
)

const listLimit = 1000

type InvoiceService struct {
	db *gorm.DB
	s3 *s3invoices.Client
}

func NewInvoiceService(db *gorm.DB, s3Client *s3invoices.Client) *InvoiceService {
	return &InvoiceService{db: db, s3: s3Client}
}

// resolveCustomer mirrors _resolve_customer: falls back to a placeholder
// when the customer record can't be found (deleted/never existed) — customers
// are soft-deleted (Active flag), never hard-deleted, so this is a
// data-integrity fallback in practice, not a normal path.
func (s *InvoiceService) resolveCustomer(ctx context.Context, inv models.Invoice) models.Customer {
	var customer models.Customer
	if err := s.db.WithContext(ctx).First(&customer, "id = ?", inv.ClienteID).Error; err == nil {
		return customer
	}
	return models.Customer{RagioneSociale: "-"}
}

func (s *InvoiceService) List(ctx context.Context, stato, clienteID string) ([]dto.InvoiceResponse, error) {
	query := s.db.WithContext(ctx).Preload("Righe").Preload("Cliente")
	if stato != "" {
		query = query.Where("stato = ?", stato)
	}
	if clienteID != "" {
		query = query.Where("cliente_id = ?", clienteID)
	}
	var invoices []models.Invoice
	if err := query.Order("created_at DESC").Limit(listLimit).Find(&invoices).Error; err != nil {
		return nil, err
	}
	result := make([]dto.InvoiceResponse, len(invoices))
	for i, inv := range invoices {
		result[i] = toResponse(inv)
	}
	return result, nil
}

// Create mirrors create_invoice: assigns a progressive "O/F-{year}/{seq}"
// number via the same shared counter used for order progressivi.
func (s *InvoiceService) Create(ctx context.Context, req dto.InvoiceRequest) (*dto.InvoiceResponse, error) {
	seq, err := database.NextSequence(s.db.WithContext(ctx), "invoices")
	if err != nil {
		return nil, err
	}
	numero := fmt.Sprintf("O/F-%s/%04d", time.Now().Format("06"), seq)

	clienteID, err := utils.ParseUUID(req.ClienteID)
	if err != nil {
		return nil, err
	}

	inv := models.Invoice{
		ID: uuid.New(), Numero: numero, ClienteID: clienteID,
		DataFattura: req.DataFattura, DataScadenza: req.DataScadenza, CondizioniPagamento: req.CondizioniPagamento,
		Righe: toLines(req.Righe), CostiAccessori: marshalJSON(req.CostiAccessori),
		TotaleImponibile: req.TotaleImponibile, TotaleIva: req.TotaleIva, Totale: req.Totale,
		Stato: "PROFORMA", Tipo: "ordine",
	}
	if err := s.db.WithContext(ctx).Create(&inv).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Preload("Righe").Preload("Cliente").First(&inv, "id = ?", inv.ID).Error; err != nil {
		return nil, err
	}
	resp := toResponse(inv)
	return &resp, nil
}

func (s *InvoiceService) GetByID(ctx context.Context, id uuid.UUID) (*dto.InvoiceResponse, error) {
	var inv models.Invoice
	if err := s.db.WithContext(ctx).Preload("Righe").Preload("Cliente").First(&inv, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	resp := toResponse(inv)
	return &resp, nil
}

// Finalize mirrors PATCH /invoices/{id}/finalize: PROFORMA -> DEFINITIVA,
// stamps fattura_id on CHIUSO orders referenced by righe[].ordine_id ("has
// this order been billed" is tracked purely via fattura_id, not a separate
// order stato — CHIUSO stays CHIUSO). The state change and order cascade
// ALWAYS happen, even if the subsequent S3 upload fails — the fiscal flow
// must not block on a transient S3 issue (matching Python's own
// comment/behavior in finalize_invoice exactly).
func (s *InvoiceService) Finalize(ctx context.Context, id uuid.UUID) (*dto.InvoiceFinalizeResult, error) {
	var inv models.Invoice
	if err := s.db.WithContext(ctx).Preload("Righe").First(&inv, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Fattura non trovata")
		}
		return nil, err
	}
	if inv.Stato != "PROFORMA" {
		return nil, utils.NewAPIError(400, "Solo fatture PROFORMA possono essere finalizzate")
	}

	if err := s.db.WithContext(ctx).Model(&models.Invoice{}).Where("id = ?", id).Update("stato", "DEFINITIVA").Error; err != nil {
		return nil, err
	}
	inv.Stato = "DEFINITIVA"

	for _, riga := range inv.Righe {
		if riga.OrdineID == "" {
			continue
		}
		s.db.WithContext(ctx).Model(&models.Order{}).
			Where("id = ? AND stato = ?", riga.OrdineID, "CHIUSO").
			Updates(map[string]interface{}{
				"fattura_id": id, "updated_at": time.Now().UTC(),
			})
	}

	result := &dto.InvoiceFinalizeResult{OK: true, PdfArchived: false, PdfS3Key: nil}
	if s.s3 == nil || !s.s3.IsEnabled() {
		return result, nil
	}

	customer := s.resolveCustomer(ctx, inv)
	pdfBytes, err := pdfgen.BuildInvoicePDF(inv, customer, nil)
	if err != nil {
		slog.Error("invoice_finalize_pdf_build_failed", "id", id, "error", err)
		return result, nil
	}
	year := inv.DataFattura
	if len(year) < 4 {
		year = time.Now().UTC().Format("2006")
	}
	key := s3invoices.BuildInvoiceKey(id.String(), inv.Numero, year)
	uploadResult, err := s.s3.UploadInvoicePDF(ctx, pdfBytes, key, map[string]string{
		"invoice_id": id.String(), "numero": inv.Numero, "cliente_id": inv.ClienteID.String(),
		"totale": fmt.Sprintf("%v", inv.Totale),
	})
	if err != nil {
		slog.Error("invoice_finalize_s3_upload_failed", "id", id, "error", err)
		return result, nil
	}
	if uploadResult == nil {
		return result, nil
	}

	if err := s.db.WithContext(ctx).Model(&models.Invoice{}).Where("id = ?", id).Updates(map[string]interface{}{
		"pdf_s3_key": uploadResult.Key, "pdf_uploaded_at": uploadResult.UploadedAt, "pdf_retain_until": uploadResult.RetainUntil,
	}).Error; err != nil {
		slog.Error("invoice_finalize_metadata_save_failed", "id", id, "error", err)
		return result, nil
	}

	result.PdfArchived = true
	result.PdfS3Key = &uploadResult.Key
	return result, nil
}

// GetPDF mirrors GET /invoices/{id}/pdf: prefers the archived S3 copy (the
// immutable source of truth) with a fallback to generating on the fly if
// the fetch fails, exactly like Python's try/except around fetch_invoice_pdf.
func (s *InvoiceService) GetPDF(ctx context.Context, id uuid.UUID) ([]byte, string, error) {
	var inv models.Invoice
	if err := s.db.WithContext(ctx).Preload("Righe").First(&inv, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", utils.NewAPIError(404, "Fattura non trovata")
		}
		return nil, "", err
	}

	customer := s.resolveCustomer(ctx, inv)
	filename := pdfgen.MakeInvoiceFilename(inv, customer)

	if inv.PdfS3Key != nil && *inv.PdfS3Key != "" && s.s3 != nil && s.s3.IsEnabled() {
		if pdfBytes, err := s.s3.FetchInvoicePDF(ctx, *inv.PdfS3Key); err == nil {
			return pdfBytes, filename, nil
		}
		slog.Error("invoice_s3_fetch_failed", "id", id, "error", "falling back to on-the-fly generation")
	}

	pdfBytes, err := pdfgen.BuildInvoicePDF(inv, customer, nil)
	if err != nil {
		return nil, "", err
	}
	return pdfBytes, filename, nil
}

// GetPDFPresignedURL mirrors GET /invoices/{id}/pdf-url: only available for
// DEFINITIVA invoices already archived on S3. The client must fall back to
// GET /pdf when this 404s.
func (s *InvoiceService) GetPDFPresignedURL(ctx context.Context, id uuid.UUID) (*dto.InvoicePDFURLResult, error) {
	var inv models.Invoice
	if err := s.db.WithContext(ctx).First(&inv, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Fattura non trovata")
		}
		return nil, err
	}
	if inv.PdfS3Key == nil || *inv.PdfS3Key == "" || s.s3 == nil || !s.s3.IsEnabled() {
		return nil, utils.NewAPIError(404, "PDF non archiviato su S3")
	}

	url, err := s.s3.GetPresignedURL(ctx, *inv.PdfS3Key, 0)
	if err != nil {
		return nil, utils.NewAPIError(500, "Errore generazione URL firmato")
	}
	return &dto.InvoicePDFURLResult{URL: url, RetainUntil: inv.PdfRetainUntil}, nil
}

// Delete mirrors DELETE /invoices/{id}: only PROFORMA invoices can be
// deleted, hard delete.
func (s *InvoiceService) Delete(ctx context.Context, id uuid.UUID) error {
	var inv models.Invoice
	if err := s.db.WithContext(ctx).First(&inv, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewAPIError(404, "Fattura non trovata")
		}
		return err
	}
	if inv.Stato != "PROFORMA" {
		return utils.NewAPIError(400, "Solo fatture PROFORMA possono essere eliminate")
	}
	return s.db.WithContext(ctx).Delete(&inv).Error
}

// ── Helpers ──────────────────────────────────────────────────────────────

func toLines(lines []dto.InvoiceLineDTO) []models.InvoiceLine {
	result := make([]models.InvoiceLine, len(lines))
	for i, l := range lines {
		result[i] = models.InvoiceLine{
			ID: uuid.New(), OrdineID: l.OrdineID, Descrizione: l.Descrizione, Prodotto: l.Prodotto,
			Peso: l.Peso, Quantita: defaultFloat(l.Quantita, 1), Tariffa: l.Tariffa, Totale: l.Totale,
			IvaCodice: defaultString(l.IvaCodice, "N8"),
		}
	}
	return result
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func defaultFloat(value, fallback float64) float64 {
	if value == 0 {
		return fallback
	}
	return value
}

func marshalJSON(v interface{}) datatypes.JSON {
	if v == nil {
		return datatypes.JSON("[]")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return datatypes.JSON("[]")
	}
	return datatypes.JSON(b)
}

func unmarshalMaps(raw datatypes.JSON) []map[string]interface{} {
	out := []map[string]interface{}{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	if out == nil {
		out = []map[string]interface{}{}
	}
	return out
}

func customerResponse(c models.Customer) *dto.CustomerResponse {
	if c.ID == uuid.Nil {
		return nil
	}
	return &dto.CustomerResponse{
		ID: c.ID, RagioneSociale: c.RagioneSociale, Indirizzo: c.Indirizzo, Citta: c.Citta,
		Cap: c.Cap, Provincia: c.Provincia, Nazione: c.Nazione, PartitaIva: c.PartitaIva,
		CodiceFiscale: c.CodiceFiscale, Telefono: c.Telefono, Email: c.Email, Pec: c.Pec,
		CondizioniPagamento: c.CondizioniPagamento, Note: c.Note, RichiedeRifOrdine: c.RichiedeRifOrdine,
		Active: c.Active, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func toResponse(inv models.Invoice) dto.InvoiceResponse {
	righe := make([]dto.InvoiceLineDTO, len(inv.Righe))
	for i, l := range inv.Righe {
		righe[i] = dto.InvoiceLineDTO{
			OrdineID: l.OrdineID, Descrizione: l.Descrizione, Prodotto: l.Prodotto,
			Peso: l.Peso, Quantita: l.Quantita, Tariffa: l.Tariffa, Totale: l.Totale, IvaCodice: l.IvaCodice,
		}
	}
	return dto.InvoiceResponse{
		ID: inv.ID, Numero: inv.Numero, ClienteID: inv.ClienteID.String(), Cliente: customerResponse(inv.Cliente),
		DataFattura: inv.DataFattura, DataScadenza: inv.DataScadenza, CondizioniPagamento: inv.CondizioniPagamento,
		Righe: righe, CostiAccessori: unmarshalMaps(inv.CostiAccessori),
		TotaleImponibile: inv.TotaleImponibile, TotaleIva: inv.TotaleIva, Totale: inv.Totale,
		Stato: inv.Stato, Tipo: inv.Tipo, Note: inv.Note,
		PdfS3Key: inv.PdfS3Key, PdfUploadedAt: inv.PdfUploadedAt, PdfRetainUntil: inv.PdfRetainUntil,
		CreatedAt: inv.CreatedAt,
	}
}
