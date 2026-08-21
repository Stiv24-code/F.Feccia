// Package inboundorders ports OrderMesh's order store and acceptance flow to
// GORM: inbound orders are transport-order drafts ingested from the mailbox
// or imported from PDFs, waiting for an operator to accept them. Dedup rule:
// one order per (ref, client), case/space-insensitive.
package inboundorders

import (
	"context"
	"errors"
	"fmt"
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

// OrderCreator creates the real TMS order a draft is converted into —
// implemented by services/orders.OrderService. Injected as a seam, like
// AcceptanceMailer, so this package does not import the orders package
// (which would close a cycle: orders is a much larger dependency and the
// conversion is the only thing inbound orders need from it).
type OrderCreator interface {
	Create(ctx context.Context, req dto.OrderRequest) (*dto.OrderResponse, error)
}

type InboundOrderService struct {
	db     *gorm.DB
	mailer AcceptanceMailer
	orders OrderCreator
}

func NewInboundOrderService(db *gorm.DB, mailer AcceptanceMailer, orderCreator OrderCreator) *InboundOrderService {
	return &InboundOrderService{db: db, mailer: mailer, orders: orderCreator}
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

// ListForClient returns a customer's own requests submitted via the
// self-service portal that are still drafts — everything not yet turned into
// a real order, whatever its acceptance status.
//
// The filter is OrderID IS NULL, deliberately not "status in (pending,
// modify)": those two are not the same set, and the difference used to be a
// hole. A request accepted by staff but not yet converted dropped out of
// this list while no models.Order existed to replace it, so from the
// customer's side the request they had sent — and been mailed a confirmation
// for — simply vanished from the portal. Keying on the conversion link
// instead means a request is visible here for exactly as long as there is
// nothing else to show it as: the moment Convert links an order, GET
// /me/orders takes over.
func (s *InboundOrderService) ListForClient(ctx context.Context, clienteID uuid.UUID) ([]dto.InboundOrderResponse, error) {
	var items []models.InboundOrder
	if err := s.db.WithContext(ctx).
		Where("cliente_id = ? AND order_id IS NULL", clienteID).
		Order("received_at DESC").
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
		ClienteID:     req.ClienteID,

		CommittenteID:         req.CommittenteID,
		DestinazioneCaricoID:  req.DestinazioneCaricoID,
		DestinazioneScaricoID: req.DestinazioneScaricoID,
		OraRitiroDa:           req.OraRitiroDa,
		OraRitiroA:            req.OraRitiroA,
		OraConsegnaDa:         req.OraConsegnaDa,
		OraConsegnaA:          req.OraConsegnaA,
		TariffaProposta:       req.TariffaProposta,
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

// Convert turns a draft into a real models.Order and links the two via
// InboundOrder.OrderID — the explicit acceptance step the type comment on
// models.InboundOrder describes. Kept separate from Accept, which only sends
// the confirmation mail and moves the status: an operator may well want to
// create the order before mailing, or mail a customer back and only enter
// the order once the anagrafica exists. Convert therefore does not touch
// Status, so "accepted" keeps meaning exactly "the confirmation mail went
// out" and OrderID keeps meaning exactly "this became an order".
//
// Which customer gets billed is the security-sensitive part. A draft from
// the mail scraper or a PDF is attacker-influenced: anyone who knows the
// scraped mailbox address can send an order naming any Client they like. So
// the free-text Client is never resolved to an anagrafica by name matching —
// that would let a stranger have orders created (and later invoiced)
// against a real customer just by spoofing a sender. ClienteID comes only
// from a trusted source:
//
//   - req.ClienteID, chosen here and now by a staff user who already holds
//     the same write role that lets them create any order outright, or
//   - the draft's own ClienteID, which only a trusted principal can have
//     set: CreateMyInboundOrder from the submitter's JWT, or a staff writer
//     on POST /inbound-orders. Neither the mail scraper nor the PDF import
//     populates it (InboundOrderDraftDTO has no such field), which is what
//     makes it safe to trust here without re-checking the source.
//
// With neither, conversion is refused (400) rather than guessed. Same
// posture for the price: Rate on a mail draft is attacker-controlled text
// and is never parsed into Order.Tariffa. The only price that carries over
// on its own is TariffaProposta, which only a portal submission can set, and
// the response flags when that is what got applied.
func (s *InboundOrderService) Convert(ctx context.Context, id uuid.UUID, req dto.InboundOrderConvertRequest) (*dto.InboundOrderConvertResponse, error) {
	if s.orders == nil {
		return nil, utils.NewAPIError(500, "conversione non disponibile: order service non configurato")
	}

	var o models.InboundOrder
	if err := s.db.WithContext(ctx).First(&o, "id = ?", id).Error; err != nil {
		return nil, err
	}
	// Idempotenza: la seconda conversione non crea un ordine gemello. Il
	// vincolo vive qui e non in un indice unico perche' order_id resta
	// legittimamente NULL su ogni draft non convertito, e piu' NULL non si
	// escludono a vicenda.
	if o.OrderID != nil {
		return nil, utils.NewAPIError(409, "richiesta gia' convertita nell'ordine "+o.OrderID.String())
	}

	clienteID := strings.TrimSpace(req.ClienteID)
	if clienteID == "" {
		if o.ClienteID == nil {
			return nil, utils.NewAPIError(400, fmt.Sprintf(
				"cliente_id obbligatorio: la richiesta arriva da %q, dove il campo client (%q) e' testo libero non riconducibile a un'anagrafica",
				o.Source, o.Client))
		}
		clienteID = o.ClienteID.String()
	}

	tariffa := o.TariffaProposta
	tariffaFromClient := o.Source == models.InboundOrderSourcePortal && o.TariffaProposta != 0
	if req.Tariffa != nil {
		tariffa = *req.Tariffa
		tariffaFromClient = false
	}

	orderReq := dto.OrderRequest{
		ClienteID:             clienteID,
		CommittenteID:         firstNonEmpty(strings.TrimSpace(req.CommittenteID), uuidString(o.CommittenteID)),
		DestinazioneCaricoID:  firstNonEmpty(strings.TrimSpace(req.DestinazioneCaricoID), uuidString(o.DestinazioneCaricoID)),
		DestinazioneScaricoID: firstNonEmpty(strings.TrimSpace(req.DestinazioneScaricoID), uuidString(o.DestinazioneScaricoID)),
		DataRitiro:            firstNonEmpty(strings.TrimSpace(req.DataRitiro), o.LoadDate),
		OraRitiroDa:           o.OraRitiroDa,
		OraRitiroA:            o.OraRitiroA,
		DataConsegna:          firstNonEmpty(strings.TrimSpace(req.DataConsegna), o.DeliveryDate),
		OraConsegnaDa:         o.OraConsegnaDa,
		OraConsegnaA:          o.OraConsegnaA,
		Tariffa:               tariffa,
		TipoTariffa:           req.TipoTariffa,
		Tipologia:             req.Tipologia,
		RifOrdineCliente:      o.Ref,
		Note:                  convertNote(o, req.Note),
	}

	created, err := s.orders.Create(ctx, orderReq)
	if err != nil {
		return nil, err
	}

	// Il link non e' nella stessa transazione del Create dell'ordine: quello
	// vive nell'orders service, sulla sua *gorm.DB. Se la scrittura del link
	// fallisce si cancella l'ordine appena creato, altrimenti resterebbe
	// orfano e riconvertibile — due ordini per una richiesta, che e' proprio
	// cio' che OrderID esiste per impedire.
	res := s.db.WithContext(ctx).Model(&models.InboundOrder{}).
		Where("id = ? AND order_id IS NULL", id).
		Update("order_id", created.ID)
	if res.Error != nil || res.RowsAffected == 0 {
		s.db.WithContext(ctx).Delete(&models.Order{}, "id = ?", created.ID)
		if res.Error != nil {
			return nil, res.Error
		}
		// RowsAffected == 0: un'altra richiesta ha convertito lo stesso
		// draft mentre creavamo l'ordine. La clausola order_id IS NULL sopra
		// e' cio' che rende la corsa innocua.
		return nil, utils.NewAPIError(409, "richiesta convertita da un'altra operazione in corso")
	}

	var reloaded models.InboundOrder
	if err := s.db.WithContext(ctx).First(&reloaded, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &dto.InboundOrderConvertResponse{
		InboundOrder:      toResponse(reloaded),
		Order:             *created,
		TariffaFromClient: tariffaFromClient,
	}, nil
}

// convertNote folds everything the draft carries that has no typed home on
// models.Order into the order note: product and kg are free text on a draft
// (the portal form asks for them as text, and OrderItem needs a real product
// FK), and the provenance line keeps the trail back to the received request
// readable to whoever opens the order later.
func convertNote(o models.InboundOrder, operatorNote string) string {
	parts := make([]string, 0, 5)
	if n := strings.TrimSpace(operatorNote); n != "" {
		parts = append(parts, n)
	}
	if p := strings.TrimSpace(o.Product); p != "" {
		parts = append(parts, "Prodotto: "+p)
	}
	if o.Kg > 0 {
		parts = append(parts, fmt.Sprintf("Kg: %d", o.Kg))
	}
	if n := strings.TrimSpace(o.Notes); n != "" {
		parts = append(parts, n)
	}
	origin := "Da richiesta " + o.Source
	if r := strings.TrimSpace(o.Ref); r != "" {
		origin += " " + r
	}
	origin += " del " + o.ReceivedAt.Format("02/01/2006")
	if rate := strings.TrimSpace(o.Rate); rate != "" {
		origin += " (tariffa indicata dal cliente: " + rate + ")"
	}
	parts = append(parts, origin)
	return strings.Join(parts, " | ")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func uuidString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
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
		ClienteID:     o.ClienteID,

		CommittenteID:         o.CommittenteID,
		DestinazioneCaricoID:  o.DestinazioneCaricoID,
		DestinazioneScaricoID: o.DestinazioneScaricoID,
		OraRitiroDa:           o.OraRitiroDa,
		OraRitiroA:            o.OraRitiroA,
		OraConsegnaDa:         o.OraConsegnaDa,
		OraConsegnaA:          o.OraConsegnaA,
		TariffaProposta:       o.TariffaProposta,
		OrderID:               o.OrderID,
	}
}
