// Package mailscraper ports OrderMesh's mailbox ingestion: it reads the
// orders mailbox — via Microsoft Graph for Exchange Online (where basic-auth
// IMAP is disabled) or classic IMAP for everything else — parses mails whose
// subject contains the [ORDINE] marker into inbound-order drafts and stores
// the new ones through the inbound-orders service.
package mailscraper

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"fratelli-feccia/config"
	"fratelli-feccia/internal/dto"
)

// OrderSink stores parsed orders, silently skipping duplicates — implemented
// by inboundorders.InboundOrderService.AddIfNew.
type OrderSink interface {
	AddIfNew(ctx context.Context, req dto.InboundOrderRequest) (bool, error)
}

type MailScraperService struct {
	cfg  config.InboundConfig
	sink OrderSink

	// app-only (client credentials) token cache — see graph.go.
	appTok struct {
		sync.Mutex
		val string
		exp time.Time
	}
}

func NewMailScraperService(cfg config.InboundConfig, sink OrderSink) *MailScraperService {
	return &MailScraperService{cfg: cfg, sink: sink}
}

// Backend resolves the configured mail backend ("imap" or "graph").
func (s *MailScraperService) Backend() string { return s.cfg.Backend() }

// MailboxReady reports whether the scraping backend can run right now:
// Graph needs app-only credentials or a saved delegated token, IMAP needs
// host+credentials in the environment.
func (s *MailScraperService) MailboxReady() bool {
	if s.cfg.Backend() == "graph" {
		return s.appOnly() || s.loggedIn()
	}
	return s.cfg.IMAPConfigured()
}

// Scrape reads the mailbox once and stores the new orders. It dispatches to
// the configured backend: Microsoft Graph for Exchange Online mailboxes,
// classic IMAP for everything else.
func (s *MailScraperService) Scrape(ctx context.Context) (added, scanned int, err error) {
	if s.cfg.Backend() == "graph" {
		return s.scrapeGraph(ctx)
	}
	return s.scrapeIMAP(ctx)
}

// Status describes the mail side of the inbound pipeline for the
// /inbound-config endpoint (the PDF flags are filled in by the handler from
// the PDF engine).
func (s *MailScraperService) Status() dto.InboundConfigResponse {
	return dto.InboundConfigResponse{
		AcceptMode:        s.cfg.AcceptMode,
		TestRecipient:     s.cfg.TestRecipient,
		SmtpReady:         s.cfg.SMTPConfigured(),
		MailboxReady:      s.MailboxReady(),
		Backend:           s.cfg.Backend(),
		SubjectFilter:     s.cfg.SubjectFilter,
		ScrapeIntervalMin: s.cfg.ScrapeIntervalMin,
	}
}

// store hands one parsed order to the sink, logging new arrivals.
func (s *MailScraperService) store(ctx context.Context, req dto.InboundOrderRequest) (added bool) {
	inserted, err := s.sink.AddIfNew(ctx, req)
	if err != nil {
		slog.Error("scrape: salvataggio ordine fallito", "ref", req.Ref, "client", req.Client, "error", err)
		return false
	}
	if inserted {
		slog.Info("scrape: nuovo ordine", "ref", req.Ref, "client", req.Client, "sender", req.SenderEmail)
	}
	return inserted
}
