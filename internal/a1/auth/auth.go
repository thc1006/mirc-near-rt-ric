package auth

import (
	"context"
	"net/http"
)

// AuthenticationHandler defines the interface for authentication and authorization
type AuthenticationHandler interface {
	AuthenticateRequest(r *http.Request) (bool, error)
	AuthorizeRequest(r *http.Request, requiredRoles []string) (bool, error)
	GenerateToken(ctx context.Context, userID string, roles []string) (string, error)
}
