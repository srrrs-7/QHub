package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"web/src/client"
	"web/src/handlers"
)

func NewRouter(apiClient client.Client) http.Handler {
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
	r.Get("/analytics/prompts/{prompt_id}", pages.PromptAnalytics())
	r.Get("/analytics/projects/{project_id}", pages.ProjectAnalytics())

	// Pages - Tags
	r.Get("/tags", pages.Tags())

	// Pages - Industries
	r.Get("/industries", pages.Industries())
	r.Get("/industries/{slug}", pages.IndustryDetail())

	// HTMX Partials
	r.Route("/partials", func(r chi.Router) {
		// Prompts & Versions
		r.Post("/projects/{project_id}/prompts", partials.CreatePrompt())
		r.Put("/projects/{project_id}/prompts/{prompt_slug}", partials.UpdatePrompt())
		r.Post("/prompts/{prompt_id}/versions", partials.CreateVersion())
		r.Get("/prompts/{prompt_id}/versions/{version}", partials.GetVersionDetail())
		r.Put("/prompts/{prompt_id}/versions/{version}/status", partials.UpdateVersionStatus())
		r.Get("/prompts/{prompt_id}/versions/{version}/lint", partials.GetLint())
		r.Get("/prompts/{prompt_id}/versions/{version}/text-diff", partials.GetTextDiff())
		r.Get("/prompts/{prompt_id}/semantic-diff/{v1}/{v2}", partials.GetSemanticDiff())
		r.Get("/prompts/{prompt_id}/compare", partials.CompareVersions())

		// Organizations
		r.Post("/organizations", partials.CreateOrganization())

		// Projects
		r.Post("/orgs/{org_id}/projects", partials.CreateProject())
		r.Put("/orgs/{org_id}/projects/{project_slug}", partials.UpdateProject())
		r.Delete("/orgs/{org_id}/projects/{project_slug}", partials.DeleteProject())

		// Evaluations
		r.Post("/logs/{log_id}/evaluations", partials.CreateEvaluation())
		r.Get("/evaluations/{id}/edit", partials.GetEvaluationEditForm())
		r.Put("/evaluations/{id}", partials.UpdateEvaluation())

		// Prompt Tags
		r.Post("/prompts/{prompt_id}/tags", partials.AddPromptTag())
		r.Delete("/prompts/{prompt_id}/tags/{tag_id}", partials.RemovePromptTag())

		// Logs
		r.Post("/logs", partials.CreateLog())

		// Noop (cancel helper for inline forms)
		r.Get("/noop", partials.Noop())

		// Consulting
		r.Post("/consulting/sessions", partials.CreateConsultingSession())
		r.Post("/consulting/sessions/{session_id}/messages", partials.SendConsultingMessage())
		r.Get("/consulting/sessions/{session_id}/stream", partials.SSEStream())

		// Settings: Organization
		r.Put("/orgs/{slug}", partials.UpdateOrganization())

		// Settings: Members
		r.Post("/orgs/{org_id}/members", partials.AddMember())
		r.Put("/orgs/{org_id}/members/{user_id}", partials.UpdateMemberRole())
		r.Delete("/orgs/{org_id}/members/{user_id}", partials.RemoveMember())

		// Settings: API Keys
		r.Post("/orgs/{org_id}/api-keys", partials.CreateAPIKey())
		r.Delete("/orgs/{org_id}/api-keys/{key_id}", partials.DeleteAPIKey())

		// Tags
		r.Post("/tags", partials.CreateTag())
		r.Delete("/tags", partials.DeleteTag())

		// Industries
		r.Post("/industries", partials.CreateIndustry())
		r.Post("/industries/{slug}/compliance", partials.CheckCompliance())

		// Analytics
		r.Get("/analytics/prompts/{prompt_id}", partials.GetPromptAnalyticsPartial())
		r.Get("/analytics/prompts/{prompt_id}/trend", partials.GetDailyTrendPartial())
		r.Get("/analytics/projects/{project_id}", partials.GetProjectAnalyticsPartial())

		// Session Close
		r.Put("/consulting/sessions/{id}/close", partials.CloseSession())

		// Embedding Status
		r.Get("/search/embedding-status", partials.GetEmbeddingStatus())

		// Search
		r.Post("/search", partials.Search())
	})

	return r
}
