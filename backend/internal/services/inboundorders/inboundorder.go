// Package inboundorders ports OrderMesh's order store and acceptance flow to
// GORM: inbound orders are transport-order drafts ingested from the mailbox
// or imported from PDFs, waiting for an operator to accept them. Dedup rule:
// one order per (ref, client), case/space-insensitive.
package inboundorders

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

// AcceptanceMailer sends the confirmation mail when an order is accepted and
// returns the actual recipient (TEST_RECIPIENT while ACCEPT_MODE=test, the
// order's sender in production). Implemented by the SMTP mailer service —
// a nil mailer means "SMTP non configurato": accept still works, no mail.
type AcceptanceMailer interface {
	SendAcceptance(ctx context.Context, o dto.InboundOrderResponse) (recipient string, err error)
}

type InboundOrderService struct {
	db     *gorm.DB
	mailer AcceptanceMailer
}

func NewInboundOrderService(db *gorm.DB, mailer AcceptanceMailer) *InboundOrderService {
	return &InboundOrderService{db: db, mailer: mailer}
}

func (s *InboundOrderService) List(ctx context.Context) ([]dto.InboundOrderResponse, error) {
	var items []models.InboundOrder
	if err := s.db.WithContext(ctx).
		Order("client ASC, received_at ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]dto.InboundOrderResponse, len(items))
	for i, o := range items {
		out[i] = toResponse(o)
	}
	return out, nil
}

// Create persists a confirmed draft. Duplicates by (ref, client) — the same
// key as the inbound_orders_ref_client_key index — answer 409, so mailbox
// re-reads and double submissions never create twin rows. The pre-check
// keeps the 409 deterministic on any database; under concurrency on
// Postgres the unique index still backstops it (mapped to 409 by
// utils.HandleDatabaseError).
func (s *InboundOrderService) Create(ctx context.Context, req dto.InboundOrderRequest) (*dto.InboundOrderResponse, error) {
	if strings.TrimSpace(req.Ref) == "" && strings.TrimSpace(req.Product) == "" {
		return nil, utils.NewAPIError(400, "ordine incompleto: servono almeno cliente e riferimento o prodotto")
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.InboundOrder{}).
		Where("lower(trim(ref)) = ? AND lower(trim(client)) = ?",
			strings.ToLower(strings.TrimSpace(req.Ref)),
			strings.ToLower(strings.TrimSpace(req.Client))).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, utils.NewAPIError(409, "ordine gia' presente (stesso riferimento e cliente)")
	}

	o := models.InboundOrder{
		ID:            uuid.New(),
		Client:        req.Client,
		SenderEmail:   strings.ToLower(strings.TrimSpace(req.SenderEmail)),
		Ref:           req.Ref,
		Product:       req.Product,
		Kg:            req.Kg,
		LoadDate:      req.LoadDate,
		LoadPlace:     req.LoadPlace,
		DeliveryDate:  req.DeliveryDate,
		DeliveryPlace: req.DeliveryPlace,
		Rate:          req.Rate,
		Notes:         req.Notes,
		Portal:        req.Portal,
		Status:        models.InboundOrderStatusPending,
		Source:        models.InboundOrderSourcePDF,
		TemplateID:    req.TemplateID,
		ReceivedAt:    time.Now(),
	}
	if req.Status != "" {
		o.Status = models.InboundOrderStatus(req.Status)
	}
	if req.Source != "" {
		o.Source = req.Source
	}
	if req.ReceivedAt != nil && !req.ReceivedAt.IsZero() {
		o.ReceivedAt = *req.ReceivedAt
	}

	if err := s.db.WithContext(ctx).Create(&o).Error; err != nil {
		return nil, err
	}
	resp := toResponse(o)
	return &resp, nil
}

// AddIfNew inserts the order unless one with the same (ref, client) already
// exists, mirroring OrderMesh's store.AddIfNew: duplicates are a normal
// outcome for the mail scraper (mailbox re-reads), not an error. Returns
// true when inserted.
func (s *InboundOrderService) AddIfNew(ctx context.Context, req dto.InboundOrderRequest) (bool, error) {
	if _, err := s.Create(ctx, req); err != nil {
		var apiErr utils.APIError
		if errors.As(err, &apiErr) && apiErr.Code == 409 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Accept sends the confirmation mail (when a mailer is configured), then
// marks the order accepted — in this order, like OrderMesh: a failed send
// leaves the order pending so the operator can retry.
func (s *InboundOrderService) Accept(ctx context.Context, id uuid.UUID) (*dto.InboundOrderActionResponse, error) {
	var o models.InboundOrder
	if err := s.db.WithContext(ctx).First(&o, "id = ?", id).Error; err != nil {
		return nil, err
	}

	mailInfo := "nessuna email inviata: SMTP non configurato"
	if s.mailer != nil {
		to, err := s.mailer.SendAcceptance(ctx, toResponse(o))
		if err != nil {
			return nil, utils.NewAPIError(502, "invio email fallito: "+err.Error())
		}
		mailInfo = "email di conferma inviata a " + to
	}

	updated, err := s.setStatus(ctx, id, models.InboundOrderStatusAccepted)
	if err != nil {
		return nil, err
	}
	return &dto.InboundOrderActionResponse{Order: *updated, Mail: mailInfo}, nil
}

// Modify marks the order as under revision; the UI then opens the local mail
// client with the sender address pre-filled via mailto:.
func (s *InboundOrderService) Modify(ctx context.Context, id uuid.UUID) (*dto.InboundOrderResponse, error) {
	return s.setStatus(ctx, id, models.InboundOrderStatusModify)
}

func (s *InboundOrderService) Reset(ctx context.Context, id uuid.UUID) (*dto.InboundOrderResponse, error) {
	return s.setStatus(ctx, id, models.InboundOrderStatusPending)
}

func (s *InboundOrderService) setStatus(ctx context.Context, id uuid.UUID, status models.InboundOrderStatus) (*dto.InboundOrderResponse, error) {
	res := s.db.WithContext(ctx).Model(&models.InboundOrder{}).
		Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var o models.InboundOrder
	if err := s.db.WithContext(ctx).First(&o, "id = ?", id).Error; err != nil {
		return nil, err
	}
	resp := toResponse(o)
	return &resp, nil
}

func toResponse(o models.InboundOrder) dto.InboundOrderResponse {
	return dto.InboundOrderResponse{
		ID:            o.ID,
		Client:        o.Client,
		SenderEmail:   o.SenderEmail,
		Ref:           o.Ref,
		Product:       o.Product,
		Kg:            o.Kg,
		LoadDate:      o.LoadDate,
		LoadPlace:     o.LoadPlace,
		DeliveryDate:  o.DeliveryDate,
		DeliveryPlace: o.DeliveryPlace,
		Rate:          o.Rate,
		Notes:         o.Notes,
		Portal:        o.Portal,
		Status:        string(o.Status),
		Source:        o.Source,
		TemplateID:    o.TemplateID,
		ReceivedAt:    o.ReceivedAt,
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
	}
}
