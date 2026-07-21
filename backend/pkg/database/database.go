package database

import (
	"context"
	"fmt"
	"log/slog"
	"fratelli-feccia/config"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to the database with retry logic
func Connect(cfg *config.Config) (*gorm.DB, error) {

	appName := "fratelli-feccia-backend"

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s application_name=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
		appName,
	)

	ctx := context.Background()
	retryConfig := utils.DatabaseConnectionRetryConfig()

	// Use retry for database connection
	var db *gorm.DB

	err := utils.Retry(ctx, func() error {
		var err error
		// Attempt connection
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return fmt.Errorf("failed to open database connection: %w", err)
		}

		// Verify connection
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get database instance: %w", err)
		}
		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}

		return nil
	}, retryConfig)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
	}

	// Get *sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance for pool configuration: %w", err)
	}

	// Connection Pool Settings (tunable via env)
	maxOpen := 20
	maxIdle := 10
	connMaxLifetime := time.Duration(60) * time.Minute
	connMaxIdleTime := time.Duration(10) * time.Minute

	sqlDB.SetMaxOpenConns(maxOpen)            // Max active connections
	sqlDB.SetMaxIdleConns(maxIdle)            // Max idle connections
	sqlDB.SetConnMaxLifetime(connMaxLifetime) // Max lifetime for a connection
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime) // Max time a connection can be idle

	slog.Info("Database connected successfully",
		slog.Int("max_open", maxOpen),
		slog.Int("max_idle", maxIdle),
		slog.Duration("conn_max_lifetime", connMaxLifetime),
		slog.Duration("conn_max_idle_time", connMaxIdleTime),
	)
	return db, nil
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.Customer{},
		&models.Destination{},
		&models.Carrier{},
		&models.Garage{},
		&models.WashStation{},
		&models.Driver{},
		&models.Product{},
		&models.VehicleType{},
		&models.AccessoryCost{},
		&models.TransportCategory{},
		&models.Country{},
		&models.Bank{},
		&models.AccountingEntry{},
		&models.DriverUnavailability{},
		&Counter{},
		&models.Order{},
		&models.OrderItem{},
		&models.Vehicle{},
		&models.GPSHistoryEntry{},
		&models.TemperatureReading{},
		&models.Trip{},
		&models.TripSegment{},
		&models.RouteCache{},
		&models.PriceList{},
		&models.PriceListItem{},
		&models.Invoice{},
		&models.InvoiceLine{},
		&models.AuditLog{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	slog.Info("Database migration completed successfully")
	return nil
}
