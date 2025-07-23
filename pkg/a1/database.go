package a1

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/hctsai1006/near-rt-ric/internal/config"
	"github.com/sirupsen/logrus"
)

// Database represents the database connection and operations
type Database struct {
	config *config.DatabaseConfig
	logger *logrus.Logger
	db     *sql.DB
}

// NewDatabase creates a new database instance
func NewDatabase(cfg *config.DatabaseConfig, logger *logrus.Logger) (*Database, error) {
	db := &Database{
		config: cfg,
		logger: logger,
	}

	connStr := cfg.GetConnectionString()

	var err error
	db.db, err = sql.Open(cfg.Driver, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Ping the database to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.logger.Info("Database connection established")
	return db, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Migrate runs database migrations
func (d *Database) Migrate() error {
	if !d.config.Migration.Enabled {
		d.logger.Info("Database migrations disabled")
		return nil
	}

	d.logger.Infof("Running database migrations from %s", d.config.Migration.Directory)

	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", d.config.Migration.Directory),
		d.config.Database,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	d.logger.Info("Database migrations completed")
	return nil
}
