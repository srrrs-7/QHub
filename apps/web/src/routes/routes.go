package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"web/src/client"
	"web/src/handlers"
)

func NewRouter(apiClient *client.APIClient) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	pages := handlers.NewPageHandler(apiClient)
	partials := handlers.NewPartialHandler(apiClient)

	// Health
	r.Get("/health", handlers.HealthHandler)

	// Pages
	r.Get("/", pages.Index())
	r.Get("/orgs/{org_slug}/projects", pages.Projects())
	r.Get("/orgs/{org_slug}/projects/{project_slug}/prompts", pages.Prompts())
	r.Get("/orgs/{org_slug}/projects/{project_slug}/prompts/{prompt_slug}", pages.PromptDetail())
	r.Get("/orgs/{org_slug}/projects/{project_slug}/prompts/{prompt_slug}/v/{version}", pages.PromptDetail())

	// HTMX Partials
	r.Route("/partials", func(r chi.Router) {
		r.Post("/projects/{project_id}/prompts", partials.CreatePrompt())
		r.Post("/prompts/{prompt_id}/versions", partials.CreateVersion())
		r.Get("/prompts/{prompt_id}/versions/{version}", partials.GetVersionDetail())
		r.Put("/prompts/{prompt_id}/versions/{version}/status", partials.UpdateVersionStatus())
	})

	return r
}
