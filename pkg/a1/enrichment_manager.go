package a1

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hctsai1006/near-rt-ric/internal/config"
	"github.com/sirupsen/logrus"
)

// EnrichmentManager manages A1 enrichment information
type EnrichmentManager struct {
	log     *logrus.Logger
	config  *config.A1Config
	db      *sql.DB
	mu      sync.RWMutex
}

// NewEnrichmentManager creates a new enrichment manager
func NewEnrichmentManager(config *config.A1Config, log *logrus.Logger, db *sql.DB) *EnrichmentManager {
	return &EnrichmentManager{
		config: config,
		log:    log,
		db:     db,
	}
}

// GetEnrichmentJobs returns all enrichment jobs
func (em *EnrichmentManager) GetEnrichmentJobs() ([]*EnrichmentInfo, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	rows, err := em.db.Query("SELECT ei_job_id, ei_type_id, job_owner, job_data, target_uri, created_at, updated_at, status FROM enrichment_jobs")
	if err != nil {
		return nil, fmt.Errorf("failed to query enrichment jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]*EnrichmentInfo, 0)
	for rows.Next() {
		job := &EnrichmentInfo{}
		if err := rows.Scan(&job.EIJobID, &job.EITypeID, &job.JobOwner, &job.JobData, &job.TargetURI, &job.CreatedAt, &job.UpdatedAt, &job.Status); err != nil {
			return nil, fmt.Errorf("failed to scan enrichment job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// DeleteEnrichmentJob deletes an enrichment job
func (em *EnrichmentManager) DeleteEnrichmentJob(jobID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Check if enrichment job exists
	var count int
	err := em.db.QueryRow("SELECT COUNT(*) FROM enrichment_jobs WHERE ei_job_id = $1", jobID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing enrichment job: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("enrichment job not found")
	}

	// Delete from database
	_, err = em.db.Exec("DELETE FROM enrichment_jobs WHERE ei_job_id = $1", jobID)
	if err != nil {
		return fmt.Errorf("failed to delete enrichment job: %w", err)
	}

	em.log.Infof("Deleted enrichment job %s", jobID)

	return nil
}
