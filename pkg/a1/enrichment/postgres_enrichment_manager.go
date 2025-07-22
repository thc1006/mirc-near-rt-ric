package enrichment

import (
	"context"
	"fmt"

	"github.com/hctsai1006/near-rt-ric/pkg/a1/db"
	"github.com/sirupsen/logrus"
)

// PostgresEnrichmentManager implements the EnrichmentManager interface using a PostgreSQL database
type PostgresEnrichmentManager struct {
	db     *db.DB
	logger *logrus.Logger
}

// NewPostgresEnrichmentManager creates a new PostgresEnrichmentManager
func NewPostgresEnrichmentManager(db *db.DB) *PostgresEnrichmentManager {
	return &PostgresEnrichmentManager{
		db:     db,
		logger: logrus.WithField("component", "postgres-enrichment-manager").Logger,
	}
}

// CreateEnrichmentInfoTable creates the enrichment_info table in the database if it doesn't exist
func (m *PostgresEnrichmentManager) CreateEnrichmentInfoTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS enrichment_info (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		source TEXT NOT NULL,
		content BYTEA NOT NULL,
		timestamp TIMESTAMPTZ NOT NULL,
		ttl BIGINT NOT NULL
	);`
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create enrichment_info table: %w", err)
	}
	m.logger.Info("Enrichment info table created or already exists")
	return nil
}

// UpdateInfo adds or updates enrichment information in the database
func (m *PostgresEnrichmentManager) UpdateInfo(ctx context.Context, info *EnrichmentInfo) error {
	query := `
	INSERT INTO enrichment_info (id, name, source, content, timestamp, ttl)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (id) DO UPDATE SET
		name = EXCLUDED.name,
		source = EXCLUDED.source,
		content = EXCLUDED.content,
		timestamp = EXCLUDED.timestamp,
		ttl = EXCLUDED.ttl;`
	_, err := m.db.ExecContext(ctx, query, info.ID, info.Name, info.Source, info.Content, info.Timestamp, info.TTL)
	if err != nil {
		return fmt.Errorf("failed to update enrichment info: %w", err)
	}
	m.logger.Infof("Successfully updated enrichment info with ID: %s", info.ID)
	return nil
}

// GetInfo retrieves enrichment information from the database by its ID
func (m *PostgresEnrichmentManager) GetInfo(ctx context.Context, id string) (*EnrichmentInfo, error) {
	query := `SELECT id, name, source, content, timestamp, ttl FROM enrichment_info WHERE id = $1;`
	row := m.db.QueryRowContext(ctx, query, id)
	info := &EnrichmentInfo{}
	err := row.Scan(&info.ID, &info.Name, &info.Source, &info.Content, &info.Timestamp, &info.TTL)
	if err != nil {
		return nil, fmt.Errorf("failed to get enrichment info with ID %s: %w", id, err)
	}
	return info, nil
}

// GetAllInfo retrieves all enrichment information from the database
func (m *PostgresEnrichmentManager) GetAllInfo(ctx context.Context) ([]*EnrichmentInfo, error) {
	query := `SELECT id, name, source, content, timestamp, ttl FROM enrichment_info;`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all enrichment info: %w", err)
	}
	defer rows.Close()

	var infoList []*EnrichmentInfo
	for rows.Next() {
		info := &EnrichmentInfo{}
		err := rows.Scan(&info.ID, &info.Name, &info.Source, &info.Content, &info.Timestamp, &info.TTL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan enrichment info row: %w", err)
		}
		infoList = append(infoList, info)
	}
	return infoList, nil
}
