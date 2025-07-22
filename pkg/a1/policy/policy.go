package policy

import "context"

// PolicyStatus represents the status of a policy
type PolicyStatus string

const (
	PolicyStatusDraft     PolicyStatus = "DRAFT"
	PolicyStatusActive    PolicyStatus = "ACTIVE"
	PolicyStatusInactive  PolicyStatus = "INACTIVE"
	PolicyStatusArchived  PolicyStatus = "ARCHIVED"
)

// Policy represents a policy in the system
type Policy struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Version   string       `json:"version"`
	Content   []byte       `json:"content"`
	Status    PolicyStatus `json:"status"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
}

// PolicyManager defines the interface for managing policies
type PolicyManager interface {
	CreatePolicy(ctx context.Context, policy *Policy) error
	GetPolicy(ctx context.Context, id string) (*Policy, error)
	UpdatePolicy(ctx context.Context, policy *Policy) error
	DeletePolicy(ctx context.Context, id string) error
	GetAllPolicies(ctx context.Context) ([]*Policy, error)
}