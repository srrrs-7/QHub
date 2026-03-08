package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"web/src/client"
	"web/src/templates"
)

type renderable interface {
	Render(ctx context.Context, w io.Writer) error
}

type PageHandler struct {
	api *client.APIClient
}

func NewPageHandler(api *client.APIClient) *PageHandler {
	return &PageHandler{api: api}
}

func render(w http.ResponseWriter, r *http.Request, c renderable) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := c.Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func (h *PageHandler) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, templates.IndexPage())
	}
}

func (h *PageHandler) Projects() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgSlug := chi.URLParam(r, "org_slug")

		org, err := h.api.GetOrganization(r.Context(), orgSlug)
		if err != nil {
			http.Error(w, "Organization not found", http.StatusNotFound)
			return
		}

		projects, err := h.api.ListProjects(r.Context(), org.ID)
		if err != nil {
			projects = []client.Project{}
		}

		page := templates.PageData{OrgSlug: orgSlug, ActiveNav: "projects"}
		render(w, r, templates.ProjectsPage(page, org, projects))
	}
}

func (h *PageHandler) Prompts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgSlug := chi.URLParam(r, "org_slug")
		projectSlug := chi.URLParam(r, "project_slug")

		org, err := h.api.GetOrganization(r.Context(), orgSlug)
		if err != nil {
			http.Error(w, "Organization not found", http.StatusNotFound)
			return
		}

		project, err := h.api.GetProject(r.Context(), org.ID, projectSlug)
		if err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		prompts, err := h.api.ListPrompts(r.Context(), project.ID)
		if err != nil {
			prompts = []client.Prompt{}
		}

		page := templates.PageData{OrgSlug: orgSlug, ProjectSlug: projectSlug, ActiveNav: "prompts"}
		render(w, r, templates.PromptsPage(page, project, prompts))
	}
}

func (h *PageHandler) PromptDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgSlug := chi.URLParam(r, "org_slug")
		projectSlug := chi.URLParam(r, "project_slug")
		promptSlug := chi.URLParam(r, "prompt_slug")

		org, err := h.api.GetOrganization(r.Context(), orgSlug)
		if err != nil {
			http.Error(w, "Organization not found", http.StatusNotFound)
			return
		}

		project, err := h.api.GetProject(r.Context(), org.ID, projectSlug)
		if err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		prompt, err := h.api.GetPrompt(r.Context(), project.ID, promptSlug)
		if err != nil {
			http.Error(w, "Prompt not found", http.StatusNotFound)
			return
		}

		versions, err := h.api.ListVersions(r.Context(), prompt.ID)
		if err != nil {
			versions = []client.PromptVersion{}
		}

		var activeVersion *client.PromptVersion
		versionParam := chi.URLParam(r, "version")
		if versionParam != "" {
			v, err := h.api.GetVersion(r.Context(), prompt.ID, versionParam)
			if err == nil {
				activeVersion = v
			}
		} else if len(versions) > 0 {
			activeVersion = &versions[0]
		}

		page := templates.PageData{OrgSlug: orgSlug, ProjectSlug: projectSlug, ActiveNav: "prompts"}
		render(w, r, templates.PromptDetailPage(page, project, prompt, versions, activeVersion))
	}
}
