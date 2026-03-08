package routes

import (
	mw "api/src/routes/middleware"
	"api/src/routes/organizations"
	"api/src/routes/projects"
	"api/src/routes/prompts"
	"api/src/routes/tasks"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handlers struct {
	Health       http.HandlerFunc
	Task         *tasks.TaskHandler
	Organization *organizations.OrganizationHandler
	Project      *projects.ProjectHandler
	Prompt       *prompts.PromptHandler
}

func NewRouter(h Handlers) http.Handler {
	r := chi.NewRouter()

	// ミドルウェア
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(mw.Logger)

	// health check
	r.Get("/health", h.Health)

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// Bearer認証を適用
			r.Use(mw.BearerAuth(validateToken))

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

			// Projects (nested under org)
			r.Route("/organizations/{org_id}/projects", func(r chi.Router) {
				r.Get("/", h.Project.List())
				r.Post("/", h.Project.Post())
				r.Get("/{project_slug}", h.Project.Get())
				r.Put("/{project_slug}", h.Project.Put())
				r.Delete("/{project_slug}", h.Project.Delete())
			})

			// Prompts (nested under project)
			r.Route("/projects/{project_id}/prompts", func(r chi.Router) {
				r.Get("/", h.Prompt.List())
				r.Post("/", h.Prompt.Post())
				r.Get("/{prompt_slug}", h.Prompt.Get())
				r.Put("/{prompt_slug}", h.Prompt.Put())
			})

			// Prompt Versions (nested under prompt)
			r.Route("/prompts/{prompt_id}/versions", func(r chi.Router) {
				r.Get("/", h.Prompt.ListVersions())
				r.Post("/", h.Prompt.PostVersion())
				r.Get("/{version}", h.Prompt.GetVersion())
				r.Put("/{version}/status", h.Prompt.PutVersionStatus())
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
