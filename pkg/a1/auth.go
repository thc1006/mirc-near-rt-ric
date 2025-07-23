package a1

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hctsai1006/near-rt-ric/internal/config"
	"github.com/sirupsen/logrus"
)

// AuthService handles JWT authentication for A1 interface
type AuthService struct {
	config     *config.A1Config
	logger     *logrus.Logger
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	redisClient *redis.Client

	// RBAC configuration
	roles       map[string]*Role
	permissions map[string]*Permission

	// Simulated user database
	users map[string]*User
}

// User represents a user in the system
type User struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Password string   `json:"-"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// Role represents a user role with associated permissions
type Role struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []string     `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
}

// Permission represents a specific permission
type Permission struct {
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserClaims represents JWT claims for A1 users
type UserClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// AuthenticatedUser represents an authenticated user context
type AuthenticatedUser struct {
	UserID      string
	Username    string
	Email       string
	Roles       []string
	Permissions []string
	Token       string
	ExpiresAt   time.Time
}

// TokenRequest represents a token request
type TokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	ClientID string `json:"client_id,omitempty"`
	Scope    string `json:"scope,omitempty"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// NewAuthService creates a new authentication service
func NewAuthService(config *config.A1Config, logger *logrus.Logger, redisClient *redis.Client) (*AuthService, error) {
	auth := &AuthService{
		config:         config,
		logger:         logger,
		redisClient:    redisClient,
		roles:          make(map[string]*Role),
		permissions:    make(map[string]*Permission),
		users:          make(map[string]*User),
	}

	// Initialize simulated user database
	auth.initializeUsers()

	// Initialize RSA keys for JWT signing
	if err := auth.loadRSAKeys(); err != nil {
		return nil, fmt.Errorf("failed to load RSA keys: %w", err)
	}

	// Initialize default roles and permissions
	auth.initializeDefaultRBAC()

	auth.logger.Info("A1 authentication service initialized")
	return auth, nil
}

// initializeUsers sets up the simulated user database
func (a *AuthService) initializeUsers() {
	a.users["admin"] = &User{
		UserID:   "admin-id",
		Username: "admin",
		Password: "password",
		Email:    "admin@example.com",
		Roles:    []string{"admin"},
	}
	a.users["operator"] = &User{
		UserID:   "operator-id",
		Username: "operator",
		Password: "password",
		Email:    "operator@example.com",
		Roles:    []string{"operator"},
	}
	a.users["viewer"] = &User{
		UserID:   "viewer-id",
		Username: "viewer",
		Password: "password",
		Email:    "viewer@example.com",
		Roles:    []string{"viewer"},
	}
}

// loadRSAKeys loads RSA keys for JWT signing from the configuration
func (a *AuthService) loadRSAKeys() error {
	if a.config.Auth.PrivateKeyPath == "" || a.config.Auth.PublicKeyPath == "" {
		return fmt.Errorf("private or public key path not specified in config")
	}

	privateKeyBytes, err := os.ReadFile(a.config.Auth.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	publicKeyBytes, err := os.ReadFile(a.config.Auth.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	a.privateKey = privateKey
	a.publicKey = publicKey

	a.logger.Info("Successfully loaded RSA keys for JWT signing")
	return nil
}

// initializeDefaultRBAC sets up default roles and permissions
func (a *AuthService) initializeDefaultRBAC() {
	// Define permissions
	permissions := []*Permission{
		{Name: "policy:read", Resource: "policy", Action: "read", Description: "Read policies"},
		{Name: "policy:write", Resource: "policy", Action: "write", Description: "Create/update policies"},
		{Name: "policy:delete", Resource: "policy", Action: "delete", Description: "Delete policies"},
		{Name: "policytype:read", Resource: "policytype", Action: "read", Description: "Read policy types"},
		{Name: "policytype:write", Resource: "policytype", Action: "write", Description: "Create/update policy types"},
		{Name: "policytype:delete", Resource: "policytype", Action: "delete", Description: "Delete policy types"},
		{Name: "enrichment:read", Resource: "enrichment", Action: "read", Description: "Read enrichment info"},
		{Name: "enrichment:write", Resource: "enrichment", Action: "write", Description: "Create/update enrichment info"},
		{Name: "model:read", Resource: "model", Action: "read", Description: "Read ML models"},
		{Name: "model:write", Resource: "model", Action: "write", Description: "Deploy ML models"},
		{Name: "admin:read", Resource: "admin", Action: "read", Description: "Read admin resources"},
		{Name: "admin:write", Resource: "admin", Action: "write", Description: "Manage admin resources"},
	}

	for _, perm := range permissions {
		perm.CreatedAt = time.Now()
		a.permissions[perm.Name] = perm
	}

	// Define roles
	roles := []*Role{
		{
			Name:        "viewer",
			Description: "Read-only access to policies and types",
			Permissions: []string{"policy:read", "policytype:read", "enrichment:read", "model:read"},
		},
		{
			Name:        "operator",
			Description: "Can manage policies but not types",
			Permissions: []string{"policy:read", "policy:write", "policy:delete", "policytype:read", "enrichment:read", "enrichment:write", "model:read"},
		},
		{
			Name:        "admin",
			Description: "Full access to all resources",
			Permissions: []string{
				"policy:read", "policy:write", "policy:delete",
				"policytype:read", "policytype:write", "policytype:delete",
				"enrichment:read", "enrichment:write",
				"model:read", "model:write",
				"admin:read", "admin:write",
			},
		},
	}

	for _, role := range roles {
		role.CreatedAt = time.Now()
		a.roles[role.Name] = role
	}

	a.logger.WithFields(logrus.Fields{
		"permissions": len(a.permissions),
		"roles":       len(a.roles),
	}).Info("Default RBAC configuration initialized")
}

// AuthenticateUser validates user credentials and returns the user details
func (a *AuthService) AuthenticateUser(username, password string) (*AuthenticatedUser, error) {
	// In a real implementation, this would query a database.
	user, exists := a.users[username]
	if !exists {
		return nil, fmt.Errorf("invalid credentials")
	}

	// In a real implementation, use a secure password hashing algorithm like bcrypt.
	if user.Password != password {
		return nil, fmt.Errorf("invalid credentials")
	}

	return &AuthenticatedUser{
		UserID:   user.UserID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    user.Roles,
	}, nil
}

// GenerateToken generates a JWT token for an authenticated user
func (a *AuthService) GenerateToken(userID, username, email string, roles []string) (*TokenResponse, error) {
	now := time.Now()
	expiresAt := now.Add(a.config.Auth.TokenExpiry)

	// Resolve permissions from roles
	permissions := a.resolvePermissions(roles)

	// Create claims
	claims := UserClaims{
		UserID:      userID,
		Username:    username,
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.config.Auth.Issuer,
			Subject:   userID,
			Audience:  []string{a.config.Auth.Audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("%s-%d", userID, now.Unix()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	
	// Sign token
	tokenString, err := token.SignedString(a.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	response := &TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   int64(a.config.Auth.TokenExpiry),
		Scope:       strings.Join(permissions, " "),
	}

	a.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"username":    username,
		"roles":       roles,
		"permissions": len(permissions),
		"expires_at":  expiresAt,
	}).Info("JWT token generated")

	return response, nil
}

// ValidateToken validates a JWT token and returns user claims
func (a *AuthService) ValidateToken(tokenString string) (*UserClaims, error) {
	// Check if token is blacklisted in Redis
	isBlacklisted, err := a.redisClient.Exists(context.Background(), tokenString).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if isBlacklisted > 0 {
		return nil, fmt.Errorf("token is blacklisted")
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Additional validation
	if claims.Issuer != a.config.Auth.Issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token is expired")
	}

	return claims, nil
}

// RevokeToken adds a token to the Redis blacklist
func (a *AuthService) RevokeToken(tokenString string) error {
	claims, err := a.ValidateToken(tokenString)
	if err != nil {
		// Even if the token is invalid, we can still add it to the blacklist
		// to prevent it from being used in the future.
		a.redisClient.Set(context.Background(), tokenString, "revoked", 24*time.Hour)
		return nil
	}

	// Add to blacklist with the remaining validity duration
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining > 0 {
		err := a.redisClient.Set(context.Background(), tokenString, "revoked", remaining).Err()
		if err != nil {
			return fmt.Errorf("failed to add token to blacklist: %w", err)
		}
	}

	a.logger.WithFields(logrus.Fields{
		"user_id":  claims.UserID,
		"username": claims.Username,
	}).Info("JWT token revoked")

	return nil
}

// resolvePermissions resolves permissions from a list of roles
func (a *AuthService) resolvePermissions(roles []string) []string {
	permissionSet := make(map[string]bool)
	
	for _, roleName := range roles {
		if role, exists := a.roles[roleName]; exists {
			for _, permission := range role.Permissions {
				permissionSet[permission] = true
			}
		}
	}

	permissions := make([]string, 0, len(permissionSet))
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}

	return permissions
}

// HasPermission checks if a user has a specific permission
func (a *AuthService) HasPermission(userPermissions []string, requiredPermission string) bool {
	for _, permission := range userPermissions {
		if permission == requiredPermission {
			return true
		}
	}
	return false
}

// AuthMiddleware returns an HTTP middleware for JWT authentication
func (a *AuthService) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				a.writeAuthError(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			// Check Bearer prefix
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				a.writeAuthError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			tokenString := authHeader[len(bearerPrefix):]
			
			// Validate token
			claims, err := a.ValidateToken(tokenString)
			if err != nil {
				a.logger.WithError(err).Warn("Token validation failed")
				a.writeAuthError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Create authenticated user context
			user := &AuthenticatedUser{
				UserID:      claims.UserID,
				Username:    claims.Username,
				Email:       claims.Email,
				Roles:       claims.Roles,
				Permissions: claims.Permissions,
				Token:       tokenString,
				ExpiresAt:   claims.ExpiresAt.Time,
			}

			// Add user to request context
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission returns middleware that checks for a specific permission
func (a *AuthService) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value("user").(*AuthenticatedUser)
			if !ok {
				a.writeAuthError(w, http.StatusUnauthorized, "User not authenticated")
				return
			}

			if !a.HasPermission(user.Permissions, permission) {
				a.logger.WithFields(logrus.Fields{
					"user_id":            user.UserID,
					"required_permission": permission,
					"user_permissions":   user.Permissions,
				}).Warn("Access denied - insufficient permissions")
				
				a.writeAuthError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeAuthError writes an authentication error response
func (a *AuthService) writeAuthError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResp := A1ErrorResponse{
		Type:   "https://tools.ietf.org/html/rfc7235#section-3.1",
		Title:  "Authentication Error",
		Status: statusCode,
		Detail: message,
	}
	
	// In a real implementation, you'd use a JSON encoder
	fmt.Fprintf(w, `{"type":"%s","title":"%s","status":%d,"detail":"%s"}`, 
		errorResp.Type, errorResp.Title, errorResp.Status, errorResp.Detail)
}

// GetUserFromContext extracts the authenticated user from request context
func GetUserFromContext(ctx context.Context) (*AuthenticatedUser, bool) {
	user, ok := ctx.Value("user").(*AuthenticatedUser)
	return user, ok
}



