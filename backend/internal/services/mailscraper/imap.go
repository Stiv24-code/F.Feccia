package mailscraper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"

	_ "github.com/emersion/go-message/charset"
)

// scrapeIMAP reads unseen messages whose subject contains the subject
// filter, parses them into orders and stores the new ones. Processed mails
// are flagged \Seen so the next run skips them.
func (s *MailScraperService) scrapeIMAP(ctx context.Context) (added, scanned int, err error) {
	cfg := s.cfg
	if !cfg.IMAPConfigured() {
		return 0, 0, fmt.Errorf("IMAP non configurato (IMAP_HOST vuoto)")
	}
	c, err := imapclient.DialTLS(cfg.IMAPHost+":"+cfg.IMAPPort, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("connessione IMAP: %w", err)
	}
	defer c.Logout()

	if err := c.Login(cfg.IMAPUser, cfg.IMAPPass).Wait(); err != nil {
		return 0, 0, fmt.Errorf("login IMAP: %w", err)
	}
	if _, err := c.Select("INBOX", nil).Wait(); err != nil {
		return 0, 0, fmt.Errorf("select INBOX: %w", err)
	}

	criteria := &imap.SearchCriteria{
		NotFlag: []imap.Flag{imap.FlagSeen},
		Header: []imap.SearchCriteriaHeaderField{
			{Key: "Subject", Value: cfg.SubjectFilter},
		},
	}
	data, err := c.Search(criteria, nil).Wait()
	if err != nil {
		return 0, 0, fmt.Errorf("search: %w", err)
	}
	nums := data.AllSeqNums()
	if len(nums) == 0 {
		return 0, 0, nil
	}

	seqSet := imap.SeqSetNum(nums...)
	fetchOptions := &imap.FetchOptions{
		Envelope:    true,
		BodySection: []*imap.FetchItemBodySection{{}},
	}
	msgs, err := c.Fetch(seqSet, fetchOptions).Collect()
	if err != nil {
		return 0, 0, fmt.Errorf("fetch: %w", err)
	}

	for _, msg := range msgs {
		scanned++
		raw := msg.FindBodySection(&imap.FetchItemBodySection{})
		if raw == nil {
			continue
		}
		fromAddr := ""
		if msg.Envelope != nil && len(msg.Envelope.From) > 0 {
			fromAddr = msg.Envelope.From[0].Addr()
		}
		body := extractTextBody(raw)
		order, ok := parseOrderBody(body, fromAddr)
		if !ok {
			slog.Info("scrape: mail ignorata (formato non riconosciuto)", "subject", msg.Envelope.Subject)
			continue
		}
		if s.store(ctx, order) {
			added++
		}
	}

	// Flag everything we examined as \Seen.
	storeCmd := c.Store(seqSet, &imap.StoreFlags{
		Op:     imap.StoreFlagsAdd,
		Silent: true,
		Flags:  []imap.Flag{imap.FlagSeen},
	}, nil)
	if err := storeCmd.Close(); err != nil {
		slog.Warn("scrape: impossibile marcare come lette", "error", err)
	}
	return added, scanned, nil
}

// extractTextBody returns the first text/plain part of a raw RFC 822 message,
// falling back to everything after the header block.
func extractTextBody(raw []byte) string {
	mr, err := mail.CreateReader(bytes.NewReader(raw))
	if err == nil {
		for {
			p, err := mr.NextPart()
			if err != nil {
				break
			}
			if _, ok := p.Header.(*mail.InlineHeader); ok {
				b, err := io.ReadAll(p.Body)
				if err == nil && len(bytes.TrimSpace(b)) > 0 {
					return string(b)
				}
			}
		}
	}
	// Naive fallback: body = everything after the first blank line.
	if _, body, found := strings.Cut(string(raw), "\r\n\r\n"); found {
		return body
	}
	if _, body, found := strings.Cut(string(raw), "\n\n"); found {
		return body
	}
	return string(raw)
}
