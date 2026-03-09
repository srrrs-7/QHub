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

	// Pages - Home
	r.Get("/", pages.Index())

	// Pages - Organizations
	r.Get("/organizations", pages.Organizations())
	r.Get("/organizations/{org_slug}", pages.OrganizationDetail())

	// Pages - Projects & Prompts (org-scoped)
	r.Get("/orgs/{org_slug}/projects", pages.Projects())
	r.Get("/orgs/{org_slug}/projects/{project_slug}/prompts", pages.Prompts())
	r.Get("/orgs/{org_slug}/projects/{project_slug}/prompts/{prompt_slug}", pages.PromptDetail())
	r.Get("/orgs/{org_slug}/projects/{project_slug}/prompts/{prompt_slug}/v/{version}", pages.PromptDetail())

	// Pages - Logs
	r.Get("/logs", pages.Logs())
	r.Get("/logs/{log_id}", pages.LogDetail())

	// Pages - Consulting
	r.Get("/consulting", pages.Consulting())
	r.Get("/consulting/{session_id}", pages.Chat())

	// Pages - Search
	r.Get("/search", pages.Search())

	// Pages - Settings
	r.Get("/orgs/{org_slug}/settings", pages.Settings())

	// Pages - Analytics
	r.Get("/analytics", pages.Analytics())

	// Pages - Tags
	r.Get("/tags", pages.Tags())

	// Pages - Industries
	r.Get("/industries", pages.Industries())
	r.Get("/industries/{slug}", pages.IndustryDetail())

	// HTMX Partials
	r.Route("/partials", func(r chi.Router) {
		// Prompts & Versions
		r.Post("/projects/{project_id}/prompts", partials.CreatePrompt())
		r.Post("/prompts/{prompt_id}/versions", partials.CreateVersion())
		r.Get("/prompts/{prompt_id}/versions/{version}", partials.GetVersionDetail())
		r.Put("/prompts/{prompt_id}/versions/{version}/status", partials.UpdateVersionStatus())

		// Organizations
		r.Post("/organizations", partials.CreateOrganization())

		// Projects
		r.Post("/orgs/{org_id}/projects", partials.CreateProject())

		// Consulting
		r.Post("/consulting/sessions", partials.CreateConsultingSession())
		r.Post("/consulting/sessions/{session_id}/messages", partials.SendConsultingMessage())
		r.Get("/consulting/sessions/{session_id}/stream", partials.SSEStream())

		// Tags
		r.Post("/tags", partials.CreateTag())
		r.Delete("/tags", partials.DeleteTag())

		// Industries
		r.Post("/industries", partials.CreateIndustry())
		r.Post("/industries/{slug}/compliance", partials.CheckCompliance())

		// Search
		r.Post("/search", partials.Search())
	})

	return r
}
