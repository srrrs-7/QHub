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
