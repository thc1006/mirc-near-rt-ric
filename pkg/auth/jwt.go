package auth

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hctsai1006/near-rt-ric/pkg/utils"
)

// Claims represents the JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates the JWT token in the Authorization header
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, you would load this from a secure source
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			if jwtSecret == "" {
				log.Fatal("JWT_SECRET environment variable not set")
			}
		}

		// Skip authentication for health check
		if r.URL.Path == "/a1-p/healthcheck" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, utils.A1ErrorUnauthorized, "Authorization header is required", nil)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, utils.A1ErrorUnauthorized, "Invalid authorization header format", nil)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				utils.WriteErrorResponse(w, http.StatusUnauthorized, utils.A1ErrorUnauthorized, "Invalid token signature", nil)
				return
			}
			utils.WriteErrorResponse(w, http.StatusUnauthorized, utils.A1ErrorUnauthorized, "Invalid token", err)
			return
		}

		if !token.Valid {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, utils.A1ErrorUnauthorized, "Invalid token", nil)
			return
		}

		// Add claims to the request context
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
