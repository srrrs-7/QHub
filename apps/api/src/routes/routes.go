package routes

import (
	mw "api/src/routes/middleware"
	"api/src/routes/tasks"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
	"utils/db/db"
	"utils/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(conn *sql.DB) http.Handler {
	r := chi.NewRouter()
	q := db.New(conn)

	// ミドルウェア
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(mw.Logger)

	// health check with DB ping
	r.Get("/health", healthHandler(conn))

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// Bearer認証を適用
			r.Use(mw.BearerAuth(validateToken))

			// Tasks
			r.Route("/tasks", func(r chi.Router) {
				r.Get("/", tasks.ListHandler(q))
				r.Post("/", tasks.PostHandler(q))
				r.Get("/{id}", tasks.GetHandler(q))
				r.Put("/{id}", tasks.PutHandler(q))
			})
		})
	})

	return r
}

// healthHandler returns a handler that checks database connectivity
func healthHandler(conn *sql.DB) http.HandlerFunc {
	type healthResponse struct {
		Status   string `json:"status"`
		Database string `json:"database"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := conn.PingContext(ctx); err != nil {
			logger.Error("health check failed", "error", err)
			dbStatus = "unhealthy"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(healthResponse{
				Status:   "unhealthy",
				Database: dbStatus,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(healthResponse{
			Status:   "ok",
			Database: dbStatus,
		})
	}
}

// validateToken validates bearer tokens
// TODO: Replace with actual JWT/Cognito validation logic
func validateToken(token string) (bool, error) {
	// Placeholder implementation - replace with actual validation
	// Example: Parse JWT, verify signature, check expiration, validate claims
	if token == "" {
		return false, nil
	}

	// TODO: Implement actual validation:
	// 1. Parse JWT token
	// 2. Verify signature using public key
	// 3. Check token expiration
	// 4. Validate issuer and audience claims
	// 5. Check token revocation if applicable

	// For development only - accept any non-empty token
	return true, nil
}
