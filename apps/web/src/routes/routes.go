package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"web/src/client"
	"web/src/handlers"
)

// NewRouter creates and configures the HTTP router.
func NewRouter(apiClient *client.APIClient) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check
	r.Get("/health", handlers.HealthHandler)

	// Main page
	r.Get("/", handlers.IndexHandler(apiClient))

	// HTMX partial endpoints
	r.Route("/partials", func(r chi.Router) {
		r.Get("/tasks", handlers.TaskListPartial(apiClient))
		r.Post("/tasks", handlers.AddTaskPartial(apiClient))
	})

	return r
}
