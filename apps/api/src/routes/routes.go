package routes

import (
	"api/src/routes/admin"
	"api/src/routes/analytics"
	"api/src/routes/apikeys"
	"api/src/routes/consulting"
	"api/src/routes/evaluations"
	"api/src/routes/industries"
	"api/src/routes/logs"
	"api/src/routes/members"
	mw "api/src/routes/middleware"
	"api/src/routes/organizations"
	"api/src/routes/projects"
	"api/src/routes/prompts"
	"api/src/routes/search"
	"api/src/routes/tags"
	"api/src/routes/tasks"
	"api/src/routes/users"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	dbq "utils/db/db"
)

type Handlers struct {
	Health       http.HandlerFunc
	Task         *tasks.TaskHandler
	Organization *organizations.OrganizationHandler
	Project      *projects.ProjectHandler
	Prompt       *prompts.PromptHandler
	Log          *logs.LogHandler
	Evaluation   *evaluations.EvaluationHandler
	Consulting   *consulting.ConsultingHandler
	Tag          *tags.TagHandler
	Industry     *industries.IndustryHandler
	Analytics    *analytics.AnalyticsHandler
	ApiKey       *apikeys.ApiKeyHandler
	Member       *members.MemberHandler
	Search       *search.SearchHandler
	User         *users.UserHandler
	Admin        *admin.AdminHandler
}

func NewRouter(h Handlers, q dbq.Querier) http.Handler {
	r := chi.NewRouter()

	// ミドルウェア
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(mw.Logger)
	r.Use(mw.NewCORSFromEnv())

	// health check
	r.Get("/health", h.Health)

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// Bearer認証を適用
			r.Use(mw.BearerAuth(validateToken))

			// テナントコンテキスト
			r.Use(mw.TenantContext())

			// レート制限
			r.Use(mw.RateLimit(mw.DefaultRateLimiterConfig()))

			// Organizations
			r.Route("/organizations", func(r chi.Router) {
				r.Post("/", h.Organization.Post())
				r.Get("/{org_slug}", h.Organization.Get())
				r.Put("/{org_slug}", h.Organization.Put())
			})

			// Tasks
			r.Route("/tasks", func(r chi.Router) {
				r.Get("/", h.Task.List())
				r.Post("/", h.Task.Post())
				r.Get("/{id}", h.Task.Get())
				r.Put("/{id}", h.Task.Put())
			})

			// Projects (nested under org) - RBAC enforced
			r.Route("/organizations/{org_id}/projects", func(r chi.Router) {
				r.With(mw.RequireRole(q, mw.RoleViewer)).Get("/", h.Project.List())
				r.With(mw.RequireRole(q, mw.RoleMember)).Post("/", h.Project.Post())
				r.With(mw.RequireRole(q, mw.RoleViewer)).Get("/{project_slug}", h.Project.Get())
				r.With(mw.RequireRole(q, mw.RoleMember)).Put("/{project_slug}", h.Project.Put())
				r.With(mw.RequireRole(q, mw.RoleAdmin)).Delete("/{project_slug}", h.Project.Delete())
			})

			// Prompts (nested under project)
			r.Route("/projects/{project_id}/prompts", func(r chi.Router) {
				r.Get("/", h.Prompt.List())
				r.Post("/", h.Prompt.Post())
				r.Get("/{prompt_slug}", h.Prompt.Get())
				r.Put("/{prompt_slug}", h.Prompt.Put())
			})

			// Prompt Versions (nested under prompt)
			r.Route("/prompts/{prompt_id}", func(r chi.Router) {
				r.Route("/versions", func(r chi.Router) {
					r.Get("/", h.Prompt.ListVersions())
					r.Post("/", h.Prompt.PostVersion())
					r.Get("/{version}", h.Prompt.GetVersion())
					r.Put("/{version}/status", h.Prompt.PutVersionStatus())
					r.Get("/{version}/lint", h.Prompt.GetLint())
					r.Get("/{version}/text-diff", h.Prompt.GetTextDiff())
					r.Get("/{version}/analytics", h.Analytics.GetVersionAnalytics())
					r.Get("/{v1}/{v2}/compare", h.Analytics.CompareVersions())
				})
				r.Get("/semantic-diff/{v1}/{v2}", h.Prompt.GetDiff())
				r.Get("/analytics", h.Analytics.GetPromptAnalytics())
				r.Get("/trend", h.Analytics.GetDailyTrend())
			})

			// Project Analytics
			r.Get("/projects/{project_id}/analytics", h.Analytics.GetProjectAnalytics())

			// Execution Logs
			r.Route("/logs", func(r chi.Router) {
				r.Get("/", h.Log.List())
				r.With(mw.BearerOrApiKeyAuth(validateToken, q)).Post("/", h.Log.Post())
				r.With(mw.BearerOrApiKeyAuth(validateToken, q)).Post("/batch", h.Log.PostBatch())
				r.Get("/{id}", h.Log.Get())
			})

			// Evaluations
			r.Route("/evaluations", func(r chi.Router) {
				r.Post("/", h.Evaluation.Post())
				r.Get("/{id}", h.Evaluation.Get())
				r.Put("/{id}", h.Evaluation.Put())
			})
			r.Get("/logs/{log_id}/evaluations", h.Evaluation.List())
			r.With(mw.BearerOrApiKeyAuth(validateToken, q)).Post("/logs/{log_id}/evaluations", h.Evaluation.Post())

			// Prompt Logs (nested under prompt)
			r.Get("/prompts/{prompt_id}/logs", h.Log.ListByPrompt())

			// Consulting Sessions
			r.Route("/consulting/sessions", func(r chi.Router) {
				r.Post("/", h.Consulting.PostSession())
				r.Get("/", h.Consulting.ListSessions())
				r.Get("/{id}", h.Consulting.GetSession())
				r.Put("/{id}", h.Consulting.PutSession())
			})

			// Consulting Messages
			r.Route("/consulting/sessions/{session_id}", func(r chi.Router) {
				r.Get("/messages", h.Consulting.ListMessages())
				r.Post("/messages", h.Consulting.PostMessage())
				r.Get("/stream", h.Consulting.Stream())
			})

			// Tags
			r.Route("/tags", func(r chi.Router) {
				r.Post("/", h.Tag.Post())
				r.Get("/", h.Tag.List())
				r.Delete("/{id}", h.Tag.Delete())
			})

			// Prompt Tags
			r.Route("/prompts/{prompt_id}/tags", func(r chi.Router) {
				r.Post("/", h.Tag.AddToPrompt())
				r.Get("/", h.Tag.ListByPrompt())
				r.Delete("/{tag_id}", h.Tag.RemoveFromPrompt())
			})

			// Industry Configs
			r.Route("/industries", func(r chi.Router) {
				r.Post("/", h.Industry.Post())
				r.Get("/", h.Industry.List())
				r.Get("/{slug}", h.Industry.GetBySlug())
				r.Put("/{slug}", h.Industry.PutBySlug())
				r.Get("/{slug}/benchmarks", h.Industry.ListBenchmarks())
				r.Post("/{slug}/compliance-check", h.Industry.ComplianceCheck())
			})

			// API Keys (nested under org) - RBAC enforced
			r.Route("/organizations/{org_id}/api-keys", func(r chi.Router) {
				r.With(mw.RequireRole(q, mw.RoleAdmin)).Post("/", h.ApiKey.Post())
				r.With(mw.RequireRole(q, mw.RoleMember)).Get("/", h.ApiKey.List())
				r.With(mw.RequireRole(q, mw.RoleAdmin)).Delete("/{id}", h.ApiKey.Delete())
			})

			// Members (nested under org) - RBAC enforced
			r.Route("/organizations/{org_id}/members", func(r chi.Router) {
				r.With(mw.RequireRole(q, mw.RoleAdmin)).Post("/", h.Member.Post())
				r.With(mw.RequireRole(q, mw.RoleMember)).Get("/", h.Member.List())
				r.With(mw.RequireRole(q, mw.RoleOwner)).Delete("/{user_id}", h.Member.Delete())
				r.With(mw.RequireRole(q, mw.RoleAdmin)).Put("/{user_id}", h.Member.Put())
			})

			// Users
			r.Route("/users", func(r chi.Router) {
				r.Post("/", h.User.Post())
				r.Get("/{id}", h.User.Get())
			})

			// Semantic Search
			r.Post("/search/semantic", h.Search.SemanticSearch())
			r.Get("/search/embedding-status", h.Search.EmbeddingStatus())

			// Admin
			r.Route("/admin", func(r chi.Router) {
				r.Post("/batch/aggregate", h.Admin.PostAggregate())
			})
		})
	})

	return r
}

// validateToken validates bearer tokens
// TODO: Replace with actual JWT/Cognito validation logic
func validateToken(token string) (bool, error) {
	if token == "" {
		return false, nil
	}

	// For development only - accept any non-empty token
	return true, nil
}
