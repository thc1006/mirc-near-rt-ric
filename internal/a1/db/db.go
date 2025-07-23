package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

// DB is a wrapper for the database connection
type DB struct {
	*sql.DB
}

// NewDB creates a new database connection
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logrus.Info("Successfully connected to the database")
	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() {
	db.DB.Close()
}
