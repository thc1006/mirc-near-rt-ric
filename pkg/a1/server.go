package a1

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/hctsai1006/near-rt-ric/pkg/a1/auth"
	"github.com/hctsai1006/near-rt-ric/pkg/a1/enrichment"
	"github.com/hctsai1006/near-rt-ric/pkg/a1/mlmodel"
	"github.com/hctsai1006/near-rt-ric/pkg/a1/policy"
)

// A1Server struct as defined in the problem description
type A1Server struct {
	policyMgr   policy.PolicyManager
	modelMgr    mlmodel.MLModelManager
	enrichMgr   enrichment.EnrichmentManager
	authHandler auth.AuthenticationHandler
	logger      *logrus.Entry
}

// NewA1Server creates a new A1Server instance
func NewA1Server(policyMgr policy.PolicyManager, modelMgr mlmodel.MLModelManager, enrichMgr enrichment.EnrichmentManager) *A1Server {
	// TODO: Move secret to a secure configuration management system
	jwtSecret := os.Getenv("A1_JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("A1_JWT_SECRET environment variable not set")
	}
	authHandler := auth.NewJWTAuthHandler(jwtSecret)

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	log := logger.WithField("component", "a1-server")

	s := &A1Server{
		policyMgr:   policyMgr,
		modelMgr:    modelMgr,
		enrichMgr:   enrichMgr,
		authHandler: authHandler,
		logger:      log,
	}
	return s
}

// Middleware functions
func (s *A1Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Infof("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (s *A1Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authenticate the request
		authenticated, err := s.authHandler.AuthenticateRequest(r)
		if !authenticated || err != nil {
			s.logger.Warnf("Authentication failed for %s %s: %v", r.Method, r.URL.Path, err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// TODO: Implement authorization based on roles and endpoint
		// For now, a simple example: require 'admin' role for POST /policies
		if r.Method == "POST" && r.URL.Path == "/a1-p/v2/policies" {
			authorized, err := s.authHandler.AuthorizeRequest(r, []string{"admin"})
			if !authorized || err != nil {
				s.logger.Warnf("Authorization failed for %s %s: %v", r.Method, r.URL.Path, err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}