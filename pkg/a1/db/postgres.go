package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

// PostgresDB represents the PostgreSQL database client.
type PostgresDB struct {
	client *sql.DB
	logger *logrus.Logger
}

// NewPostgresDB creates a new PostgreSQL database client.
func NewPostgresDB(connStr string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)                 // Max number of open connections
	db.SetMaxIdleConns(10)                 // Max number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Max lifetime for a connection

	// Ping the database to verify connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger := logrus.WithField("component", "postgres-db").Logger
	logger.Info("Successfully connected to PostgreSQL database")

	return &PostgresDB{
		client: db,
		logger: logger,
	}, nil
}

// Close closes the database connection.
func (p *PostgresDB) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// GetClient returns the underlying *sql.DB client.
func (p *PostgresDB) GetClient() *sql.DB {
	return p.client
}
