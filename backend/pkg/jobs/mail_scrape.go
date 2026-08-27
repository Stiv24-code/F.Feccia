package jobs

import (
	"context"
	"log/slog"
	"time"
)

// Scraper is the slice of the mail scraper this job needs — implemented by
// internal/services/mailscraper.MailScraperService.
type Scraper interface {
	Scrape(ctx context.Context) (added, scanned int, err error)
}

// scrapeRunTimeout bounds a single scrape run so a hung mail server (Graph
// or IMAP) can't wedge the ticker loop forever — well under the minimum
// ScrapeIntervalMin (5 minutes, see config.go) so a stuck run never bleeds
// into the next tick.
const scrapeRunTimeout = 2 * time.Minute

// StartMailScrapeJob starts the periodic mailbox read (OrderMesh's automatic
// scrape), following the same ticker/shutdown pattern as the other jobs.
// The caller decides whether to start it at all (mailbox configured and
// authenticated) — this function only paces and logs.
func StartMailScrapeJob(ctx context.Context, scraper Scraper, interval time.Duration) {
	go runMailScrapeJob(ctx, scraper, interval)
}

func runMailScrapeJob(ctx context.Context, scraper Scraper, interval time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic recovered in mail scrape job", "error", r)
		}
	}()

	slog.Info("Mail scrape job started", "interval", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Mail scrape job stopped", "reason", ctx.Err())
			return
		case <-ticker.C:
			runCtx, cancel := context.WithTimeout(ctx, scrapeRunTimeout)
			added, scanned, err := scraper.Scrape(runCtx)
			cancel()
			if err != nil {
				slog.Error("scrape periodico fallito", "error", err)
				continue
			}
			slog.Info("scrape periodico completato", "scanned", scanned, "added", added)
		}
	}
}
