package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

// JWTAuthHandler implements AuthenticationHandler using JWT
type JWTAuthHandler struct {
	secretKey []byte
	logger    *logrus.Logger
}

// NewJWTAuthHandler creates a new JWTAuthHandler
func NewJWTAuthHandler(secret string) *JWTAuthHandler {
	return &JWTAuthHandler{
		secretKey: []byte(secret),
		logger:    logrus.WithField("component", "jwt-auth").Logger,
	}
}

// CustomClaims defines the structure of the JWT claims
type CustomClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT for a given user ID and roles
func (h *JWTAuthHandler) GenerateToken(ctx context.Context, userID string, roles []string) (string, error) {
	h.logger.Infof("Generating token for userID: %s, roles: %v", userID, roles)

	claims := CustomClaims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(h.secretKey)
	if err != nil {
		h.logger.Errorf("Failed to sign token: %v", err)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// AuthenticateRequest validates the JWT from the request header
func (h *JWTAuthHandler) AuthenticateRequest(r *http.Request) (bool, error) {
	h.logger.Info("Executing JWT authentication")
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false, fmt.Errorf("missing Authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return false, fmt.Errorf("invalid token format: must be Bearer token")
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.secretKey, nil
	})

	if err != nil {
		h.logger.Warnf("Token validation failed: %v", err)
		return false, fmt.Errorf("token validation failed: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		h.logger.Infof("Successfully authenticated user: %s", claims.Subject)
		// Add claims to request context for use in handlers
		ctx := context.WithValue(r.Context(), "userClaims", claims)
		*r = *r.WithContext(ctx)
		return true, nil
	}

	return false, fmt.Errorf("invalid token")
}

// AuthorizeRequest checks if the user has the required roles
func (h *JWTAuthHandler) AuthorizeRequest(r *http.Request, requiredRoles []string) (bool, error) {
	h.logger.Infof("Executing authorization for roles: %v", requiredRoles)
	claims, ok := r.Context().Value("userClaims").(*CustomClaims)
	if !ok {
		return false, fmt.Errorf("no user claims found in context")
	}

	userRoles := claims.Roles
	for _, requiredRole := range requiredRoles {
		found := false
		for _, userRole := range userRoles {
			if userRole == requiredRole {
				found = true
				break
			}
		}
		if !found {
			h.logger.Warnf("Authorization failed: user %s does not have required role '%s'", claims.Subject, requiredRole)
			return false, fmt.Errorf("user does not have required role: %s", requiredRole)
		}
	}

	h.logger.Infof("Successfully authorized user %s for roles %v", claims.Subject, requiredRoles)
	return true, nil
}
