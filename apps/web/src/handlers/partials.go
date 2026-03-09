package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"web/src/client"
	"web/src/templates"
)

type PartialHandler struct {
	api *client.APIClient
}

func NewPartialHandler(api *client.APIClient) *PartialHandler {
	return &PartialHandler{api: api}
}

// CreatePrompt handles HTMX POST to create a prompt and return updated list.
func (h *PartialHandler) CreatePrompt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"name":        r.FormValue("name"),
			"slug":        r.FormValue("slug"),
			"prompt_type": r.FormValue("prompt_type"),
			"description": r.FormValue("description"),
		}

		if _, err := h.api.CreatePrompt(r.Context(), projectID, body); err != nil {
			renderSnackbar(w, r, "Error creating prompt: "+err.Error(), true)
			return
		}

		prompts, err := h.api.ListPrompts(r.Context(), projectID)
		if err != nil {
			prompts = []client.Prompt{}
		}

		// Re-render the prompt list (need page data from referer or just use empty for partials)
		page := templates.PageData{}
		render(w, r, templates.PromptList(page, prompts))
	}
}

// CreateVersion handles HTMX POST to create a version and return updated version list.
func (h *PartialHandler) CreateVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		content := r.FormValue("content")
		contentJSON, _ := json.Marshal(content)

		body := map[string]any{
			"content":            json.RawMessage(contentJSON),
			"change_description": r.FormValue("change_description"),
		}

		if _, err := h.api.CreateVersion(r.Context(), promptID, body); err != nil {
			renderSnackbar(w, r, "Error creating version: "+err.Error(), true)
			return
		}

		versions, err := h.api.ListVersions(r.Context(), promptID)
		if err != nil {
			versions = []client.PromptVersion{}
		}

		prompt := &client.Prompt{ID: promptID}
		page := templates.PageData{}
		for _, v := range versions {
			render(w, r, templates.VersionItem(page, prompt, v, v.VersionNumber == versions[0].VersionNumber))
		}
	}
}

// GetVersionDetail handles HTMX GET for version detail panel.
func (h *PartialHandler) GetVersionDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")
		versionNum := chi.URLParam(r, "version")

		v, err := h.api.GetVersion(r.Context(), promptID, versionNum)
		if err != nil {
			http.Error(w, "Version not found", http.StatusNotFound)
			return
		}

		prompt := &client.Prompt{ID: promptID}
		page := templates.PageData{}
		render(w, r, templates.VersionDetail(page, prompt, *v))
	}
}

// UpdateVersionStatus handles HTMX PUT for version status change.
func (h *PartialHandler) UpdateVersionStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")
		versionNum := chi.URLParam(r, "version")

		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		v, err := h.api.UpdateVersionStatus(r.Context(), promptID, versionNum, body.Status)
		if err != nil {
			renderSnackbar(w, r, "Error updating status: "+err.Error(), true)
			return
		}

		prompt := &client.Prompt{ID: promptID}
		page := templates.PageData{}
		render(w, r, templates.VersionDetail(page, prompt, *v))
	}
}

// --- Organizations ---

func (h *PartialHandler) CreateOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"name": r.FormValue("name"),
			"slug": r.FormValue("slug"),
			"plan": r.FormValue("plan"),
		}

		if _, err := h.api.CreateOrganization(r.Context(), body); err != nil {
			renderSnackbar(w, r, "Error creating organization: "+err.Error(), true)
			return
		}

		orgs, err := h.api.ListOrganizations(r.Context())
		if err != nil {
			orgs = []client.Organization{}
		}

		render(w, r, templates.OrganizationList(orgs))
	}
}

// --- Projects (from org detail) ---

func (h *PartialHandler) CreateProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "org_id")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"name":        r.FormValue("name"),
			"slug":        r.FormValue("slug"),
			"description": r.FormValue("description"),
		}

		if _, err := h.api.CreateProject(r.Context(), orgID, body); err != nil {
			renderSnackbar(w, r, "Error creating project: "+err.Error(), true)
			return
		}

		renderSnackbar(w, r, "Project created successfully", false)
	}
}

// --- Consulting ---

func (h *PartialHandler) CreateConsultingSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"title": r.FormValue("title"),
		}

		if _, err := h.api.CreateConsultingSession(r.Context(), body); err != nil {
			renderSnackbar(w, r, "Error creating session: "+err.Error(), true)
			return
		}

		sessions, err := h.api.ListConsultingSessions(r.Context())
		if err != nil {
			sessions = []client.ConsultingSession{}
		}

		render(w, r, templates.SessionList(sessions))
	}
}

func (h *PartialHandler) SendConsultingMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "session_id")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"content": r.FormValue("content"),
			"role":    "user",
		}

		msg, err := h.api.SendConsultingMessage(r.Context(), sessionID, body)
		if err != nil {
			renderSnackbar(w, r, "Error sending message: "+err.Error(), true)
			return
		}

		render(w, r, templates.ChatMessage(*msg))
	}
}

// SSEStream provides a Server-Sent Events endpoint for chat streaming.
func (h *PartialHandler) SSEStream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Keep connection alive until client disconnects
		<-r.Context().Done()
	}
}

// --- Tags ---

func (h *PartialHandler) CreateTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"name": r.FormValue("name"),
		}

		if _, err := h.api.CreateTag(r.Context(), body); err != nil {
			renderSnackbar(w, r, "Error creating tag: "+err.Error(), true)
			return
		}

		tags, err := h.api.ListTags(r.Context())
		if err != nil {
			tags = []client.Tag{}
		}

		render(w, r, templates.TagList(tags))
	}
}

func (h *PartialHandler) DeleteTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")

		if err := h.api.DeleteTag(r.Context(), name); err != nil {
			renderSnackbar(w, r, "Error deleting tag: "+err.Error(), true)
			return
		}

		tags, err := h.api.ListTags(r.Context())
		if err != nil {
			tags = []client.Tag{}
		}

		render(w, r, templates.TagList(tags))
	}
}

// --- Industries ---

func (h *PartialHandler) CreateIndustry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"name":        r.FormValue("name"),
			"slug":        r.FormValue("slug"),
			"description": r.FormValue("description"),
		}

		if _, err := h.api.CreateIndustry(r.Context(), body); err != nil {
			renderSnackbar(w, r, "Error creating industry: "+err.Error(), true)
			return
		}

		industries, err := h.api.ListIndustries(r.Context())
		if err != nil {
			industries = []client.Industry{}
		}

		render(w, r, templates.IndustryList(industries))
	}
}

func (h *PartialHandler) CheckCompliance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"content": r.FormValue("content"),
		}

		result, err := h.api.CheckCompliance(r.Context(), slug, body)
		if err != nil {
			renderSnackbar(w, r, "Error checking compliance: "+err.Error(), true)
			return
		}

		render(w, r, templates.ComplianceResultPartial(result))
	}
}

func renderSnackbar(w http.ResponseWriter, _ *http.Request, msg string, isError bool) {
	cls := "snackbar--success"
	if isError {
		cls = "snackbar--error"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<div class="snackbar %s" hx-swap-oob="innerHTML:#snackbar">%s</div>`, cls, msg)
}

// HealthHandler returns a simple health check response.
func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
