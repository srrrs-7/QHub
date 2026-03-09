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

// --- Organizations ---

func (h *PageHandler) Organizations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgs, err := h.api.ListOrganizations(r.Context())
		if err != nil {
			orgs = []client.Organization{}
		}

		page := templates.PageData{ActiveNav: "orgs"}
		render(w, r, templates.OrganizationsPage(page, orgs))
	}
}

func (h *PageHandler) OrganizationDetail() http.HandlerFunc {
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

		page := templates.PageData{OrgSlug: orgSlug, ActiveNav: "orgs"}
		render(w, r, templates.OrganizationDetailPage(page, org, projects))
	}
}

// --- Logs ---

func (h *PageHandler) Logs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logs, err := h.api.ListLogs(r.Context())
		if err != nil {
			logs = []client.ExecutionLog{}
		}

		page := templates.PageData{ActiveNav: "logs"}
		render(w, r, templates.LogsPage(page, logs))
	}
}

func (h *PageHandler) LogDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logID := chi.URLParam(r, "log_id")

		log, err := h.api.GetLog(r.Context(), logID)
		if err != nil {
			http.Error(w, "Log not found", http.StatusNotFound)
			return
		}

		evals, err := h.api.ListLogEvaluations(r.Context(), logID)
		if err != nil {
			evals = []client.Evaluation{}
		}

		page := templates.PageData{ActiveNav: "logs"}
		render(w, r, templates.LogDetailPage(page, log, evals))
	}
}

// --- Consulting ---

func (h *PageHandler) Consulting() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessions, err := h.api.ListConsultingSessions(r.Context())
		if err != nil {
			sessions = []client.ConsultingSession{}
		}

		page := templates.PageData{ActiveNav: "consulting"}
		render(w, r, templates.ConsultingPage(page, sessions))
	}
}

func (h *PageHandler) Chat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "session_id")

		session, err := h.api.GetConsultingSession(r.Context(), sessionID)
		if err != nil {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}

		messages, err := h.api.ListConsultingMessages(r.Context(), sessionID)
		if err != nil {
			messages = []client.ConsultingMessage{}
		}

		page := templates.PageData{ActiveNav: "consulting"}
		render(w, r, templates.ChatPage(page, session, messages))
	}
}

// --- Analytics ---

func (h *PageHandler) Analytics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logs, err := h.api.ListLogs(r.Context())
		if err != nil {
			logs = []client.ExecutionLog{}
		}

		evals, err := h.api.ListEvaluations(r.Context())
		if err != nil {
			evals = []client.Evaluation{}
		}

		sessions, err := h.api.ListConsultingSessions(r.Context())
		if err != nil {
			sessions = []client.ConsultingSession{}
		}

		data := templates.AnalyticsDataFromLogs(logs, evals, sessions)
		page := templates.PageData{ActiveNav: "analytics"}
		render(w, r, templates.AnalyticsPage(page, data))
	}
}

// --- Tags ---

func (h *PageHandler) Tags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tags, err := h.api.ListTags(r.Context())
		if err != nil {
			tags = []client.Tag{}
		}

		page := templates.PageData{ActiveNav: "tags"}
		render(w, r, templates.TagsPage(page, tags))
	}
}

// --- Industries ---

func (h *PageHandler) Industries() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		industries, err := h.api.ListIndustries(r.Context())
		if err != nil {
			industries = []client.Industry{}
		}

		page := templates.PageData{ActiveNav: "industries"}
		render(w, r, templates.IndustriesPage(page, industries))
	}
}

func (h *PageHandler) IndustryDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")

		industry, err := h.api.GetIndustry(r.Context(), slug)
		if err != nil {
			http.Error(w, "Industry not found", http.StatusNotFound)
			return
		}

		benchmarks, err := h.api.ListBenchmarks(r.Context(), slug)
		if err != nil {
			benchmarks = []client.Benchmark{}
		}

		page := templates.PageData{ActiveNav: "industries"}
		render(w, r, templates.IndustryDetailPage(page, industry, benchmarks))
	}
}
