package app

import (
	"log/slog"
	"os"
	"strings"

	"fratelli-feccia/internal/models"
	"fratelli-feccia/internal/seeddemo"
	"fratelli-feccia/pkg/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// seedSuperAdmin bootstraps a fresh, empty database on startup.
//
// If IS_LOCAL=true, it runs the full realistic demo dataset (seeddemo.Seed —
// same one behind `make seed-demo`): customers, fleet, orders, invoices, etc.
// This is gated behind IS_LOCAL specifically so it can never fire against a
// real deployment on its first boot — seeding fake companies (VOG, Parmalat,
// Barilla...) into a production DB would be a real mess to clean up.
//
// Otherwise (or if the demo seed fails) it falls back to a single bare admin
// user, matching this function's original behavior.
func seedSuperAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	if strings.EqualFold(os.Getenv("IS_LOCAL"), "true") {
		if err := seeddemo.Seed(db); err != nil {
			slog.Error("Failed to seed demo data, falling back to bare admin", "error", err)
		} else {
			slog.Info("Demo data seeded")
			return
		}
	}

	seedBareAdmin(db)
}

func seedBareAdmin(db *gorm.DB) {
	login := os.Getenv("SEED_ADMIN_EMAIL")
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	if login == "" {
		login = "admin"
	}
	if password == "" {
		password = "admin123"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to hash seed admin password", "error", err)
		return
	}

	user := models.User{
		Login:        login,
		Name:         "Admin",
		PasswordHash: string(hash),
		Role:         utils.RoleAdmin,
		Active:       true,
	}

	if err := db.Create(&user).Error; err != nil {
		slog.Error("Failed to seed admin user", "error", err)
		return
	}

	slog.Info("Admin user created", "login", login)
}
