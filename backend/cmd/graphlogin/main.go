// Command graphlogin runs the one-time Microsoft 365 sign-in for the mail
// scraper's delegated Graph mode, persisting the tokens in
// DATA_DIR/graph_token.json (a mounted volume in Docker).
//
// Usage:
//
//	go run ./cmd/graphlogin           # browser flow (auth-code + PKCE)
//	go run ./cmd/graphlogin -device   # device-code flow, where allowed
//
// Not needed in app-only mode (GRAPH_CLIENT_SECRET set): there the service
// authenticates as the application itself, without any user sign-in.
//
// Reads the same .env/environment as the backend — config.Load() exits when
// the backend's required variables (DB password, JWT secrets) are missing,
// so run it where the backend's env is available.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"fratelli-feccia/config"
	"fratelli-feccia/internal/services/mailscraper"
)

func main() {
	device := flag.Bool("device", false, "usa il flusso device code invece del browser")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()

	// No database needed for the login itself: the scraper is constructed
	// without an order sink.
	scraper := mailscraper.NewMailScraperService(cfg.Inbound, nil)

	var err error
	if *device {
		err = scraper.LoginDevice()
	} else {
		err = scraper.LoginBrowser()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "login:", err)
		os.Exit(1)
	}
}
