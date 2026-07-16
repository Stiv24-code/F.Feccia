// Command seed runs seeddemo.Seed manually.
//
// Usage:
//
//	SEED_ADMIN_PASSWORD="my-strong-password" go run ./cmd/seed
//
// Exits 0 on success, 1 if an admin with SEED_ADMIN_EMAIL already exists
// (refuses to reseed a non-empty DB), 1 on any other fatal error (config
// load, DB connection — config.Load() itself already os.Exit(1)s on missing
// required env vars, matching the rest of this backend's startup behavior).
package main

import (
	"fmt"
	"log/slog"
	"os"

	"fratelli-feccia/config"
	"fratelli-feccia/internal/seeddemo"
	"fratelli-feccia/pkg/database"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	if err := database.Migrate(db); err != nil {
		slog.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	if err := seeddemo.Seed(db); err != nil {
		fmt.Fprintln(os.Stderr, "ERRORE:", err)
		os.Exit(1)
	}
}
