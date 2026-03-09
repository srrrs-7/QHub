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

// --- Projects (update/delete) ---

func (h *PartialHandler) UpdateProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "org_id")
		projectSlug := chi.URLParam(r, "project_slug")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"name":        r.FormValue("name"),
			"description": r.FormValue("description"),
		}

		if _, err := h.api.UpdateProject(r.Context(), orgID, projectSlug, body); err != nil {
			renderSnackbar(w, r, "Error updating project: "+err.Error(), true)
			return
		}

		org, err := h.api.GetOrganization(r.Context(), orgID)
		if err != nil {
			renderSnackbar(w, r, "Error fetching organization: "+err.Error(), true)
			return
		}

		projects, err := h.api.ListProjects(r.Context(), orgID)
		if err != nil {
			projects = []client.Project{}
		}

		page := templates.PageData{}
		render(w, r, templates.ProjectList(page, org, projects))
	}
}

func (h *PartialHandler) DeleteProject() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "org_id")
		projectSlug := chi.URLParam(r, "project_slug")

		if err := h.api.DeleteProject(r.Context(), orgID, projectSlug); err != nil {
			renderSnackbar(w, r, "Error deleting project: "+err.Error(), true)
			return
		}

		org, err := h.api.GetOrganization(r.Context(), orgID)
		if err != nil {
			renderSnackbar(w, r, "Error fetching organization: "+err.Error(), true)
			return
		}

		projects, err := h.api.ListProjects(r.Context(), orgID)
		if err != nil {
			projects = []client.Project{}
		}

		page := templates.PageData{}
		render(w, r, templates.ProjectList(page, org, projects))
	}
}

// --- Prompts (update) ---

func (h *PartialHandler) UpdatePrompt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		promptSlug := chi.URLParam(r, "prompt_slug")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"name":        r.FormValue("name"),
			"description": r.FormValue("description"),
		}

		prompt, err := h.api.UpdatePrompt(r.Context(), projectID, promptSlug, body)
		if err != nil {
			renderSnackbar(w, r, "Error updating prompt: "+err.Error(), true)
			return
		}

		render(w, r, templates.PromptHeaderUpdated(prompt))
	}
}

// --- Version Comparison ---

func (h *PartialHandler) CompareVersions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")
		v1 := r.URL.Query().Get("v1")
		v2 := r.URL.Query().Get("v2")

		if v1 == "" || v2 == "" {
			http.Error(w, "v1 and v2 query params are required", http.StatusBadRequest)
			return
		}

		result, err := h.api.CompareVersions(r.Context(), promptID, v1, v2)
		if err != nil {
			renderSnackbar(w, r, "Error comparing versions: "+err.Error(), true)
			return
		}

		render(w, r, templates.VersionCompareCard(result))
	}
}

// --- Semantic Diff ---

func (h *PartialHandler) GetSemanticDiff() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")
		v1 := chi.URLParam(r, "v1")
		v2 := chi.URLParam(r, "v2")

		result, err := h.api.GetSemanticDiff(r.Context(), promptID, v1, v2)
		if err != nil {
			renderSnackbar(w, r, "Error loading semantic diff: "+err.Error(), true)
			return
		}

		render(w, r, templates.SemanticDiffCard(result))
	}
}

// --- Evaluations ---

func (h *PartialHandler) CreateEvaluation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logID := chi.URLParam(r, "log_id")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]any{
			"evaluator_type": r.FormValue("evaluator_type"),
		}

		if v := r.FormValue("overall_score"); v != "" {
			body["overall_score"] = v
		}
		if v := r.FormValue("accuracy_score"); v != "" {
			body["accuracy_score"] = v
		}
		if v := r.FormValue("relevance_score"); v != "" {
			body["relevance_score"] = v
		}
		if v := r.FormValue("fluency_score"); v != "" {
			body["fluency_score"] = v
		}
		if v := r.FormValue("safety_score"); v != "" {
			body["safety_score"] = v
		}
		if v := r.FormValue("feedback"); v != "" {
			body["feedback"] = v
		}

		if _, err := h.api.CreateEvaluation(r.Context(), logID, body); err != nil {
			renderSnackbar(w, r, "Error creating evaluation: "+err.Error(), true)
			return
		}

		evals, err := h.api.ListLogEvaluations(r.Context(), logID)
		if err != nil {
			evals = []client.Evaluation{}
		}

		render(w, r, templates.EvaluationsList(evals))
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

// --- Search ---

func (h *PartialHandler) Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]any{
			"query":  r.FormValue("query"),
			"org_id": r.FormValue("org_id"),
			"limit":  10,
		}

		results, err := h.api.SemanticSearch(r.Context(), body)
		if err != nil {
			renderSnackbar(w, r, "Search failed: "+err.Error(), true)
			return
		}

		render(w, r, templates.SearchResults(results))
	}
}

// --- Lint ---

func (h *PartialHandler) GetLint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")
		version := chi.URLParam(r, "version")

		result, err := h.api.GetLintResult(r.Context(), promptID, version)
		if err != nil {
			renderSnackbar(w, r, "Error loading lint: "+err.Error(), true)
			return
		}
		render(w, r, templates.LintResultCard(result))
	}
}

// --- Text Diff ---

func (h *PartialHandler) GetTextDiff() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")
		version := chi.URLParam(r, "version")

		result, err := h.api.GetTextDiff(r.Context(), promptID, version)
		if err != nil {
			renderSnackbar(w, r, "Error loading diff: "+err.Error(), true)
			return
		}
		render(w, r, templates.TextDiffCard(result))
	}
}

// --- Settings: Organization ---

func (h *PartialHandler) UpdateOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"name": r.FormValue("name"),
			"plan": r.FormValue("plan"),
		}

		org, err := h.api.UpdateOrganization(r.Context(), slug, body)
		if err != nil {
			renderSnackbar(w, r, "Error updating organization: "+err.Error(), true)
			return
		}

		render(w, r, templates.OrgInfoSection(org))
	}
}

// --- Settings: Members ---

func (h *PartialHandler) AddMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "org_id")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"user_id": r.FormValue("user_id"),
			"role":    r.FormValue("role"),
		}

		if _, err := h.api.AddMember(r.Context(), orgID, body); err != nil {
			renderSnackbar(w, r, "Error adding member: "+err.Error(), true)
			return
		}

		members, err := h.api.ListMembers(r.Context(), orgID)
		if err != nil {
			members = []client.Member{}
		}

		render(w, r, templates.MemberList(orgID, members))
	}
}

func (h *PartialHandler) UpdateMemberRole() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "org_id")
		userID := chi.URLParam(r, "user_id")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"role": r.FormValue("role"),
		}

		if _, err := h.api.UpdateMemberRole(r.Context(), orgID, userID, body); err != nil {
			renderSnackbar(w, r, "Error updating member role: "+err.Error(), true)
			return
		}

		members, err := h.api.ListMembers(r.Context(), orgID)
		if err != nil {
			members = []client.Member{}
		}

		render(w, r, templates.MemberList(orgID, members))
	}
}

func (h *PartialHandler) RemoveMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "org_id")
		userID := chi.URLParam(r, "user_id")

		if err := h.api.RemoveMember(r.Context(), orgID, userID); err != nil {
			renderSnackbar(w, r, "Error removing member: "+err.Error(), true)
			return
		}

		members, err := h.api.ListMembers(r.Context(), orgID)
		if err != nil {
			members = []client.Member{}
		}

		render(w, r, templates.MemberList(orgID, members))
	}
}

// --- Settings: API Keys ---

func (h *PartialHandler) CreateAPIKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "org_id")

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		body := map[string]string{
			"name": r.FormValue("name"),
		}

		created, err := h.api.CreateAPIKey(r.Context(), orgID, body)
		if err != nil {
			renderSnackbar(w, r, "Error creating API key: "+err.Error(), true)
			return
		}

		// Show the created key notice first (key is only shown once)
		render(w, r, templates.APIKeyCreatedNotice(created))

		// Then render the updated list
		keys, err := h.api.ListAPIKeys(r.Context(), orgID)
		if err != nil {
			keys = []client.APIKey{}
		}

		render(w, r, templates.APIKeyList(orgID, keys))
	}
}

func (h *PartialHandler) DeleteAPIKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "org_id")
		keyID := chi.URLParam(r, "key_id")

		if err := h.api.DeleteAPIKey(r.Context(), orgID, keyID); err != nil {
			renderSnackbar(w, r, "Error revoking API key: "+err.Error(), true)
			return
		}

		keys, err := h.api.ListAPIKeys(r.Context(), orgID)
		if err != nil {
			keys = []client.APIKey{}
		}

		render(w, r, templates.APIKeyList(orgID, keys))
	}
}

// --- Analytics Partials ---

// GetPromptAnalyticsPartial fetches and renders prompt analytics data.
func (h *PartialHandler) GetPromptAnalyticsPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")

		versions, err := h.api.GetPromptAnalytics(r.Context(), promptID)
		if err != nil {
			versions = []client.PromptAnalytics{}
		}

		trend, err := h.api.GetDailyTrend(r.Context(), promptID, "30")
		if err != nil {
			trend = []client.DailyTrend{}
		}

		data := templates.PromptAnalyticsData{
			PromptID:   promptID,
			PromptName: "Prompt",
			Versions:   versions,
			Trend:      trend,
			Days:       30,
		}

		render(w, r, templates.PromptAnalyticsPartial(data))
	}
}

// GetDailyTrendPartial fetches and renders daily trend data with a configurable days parameter.
func (h *PartialHandler) GetDailyTrendPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")
		days := r.URL.Query().Get("days")
		if days == "" {
			days = "30"
		}

		trend, err := h.api.GetDailyTrend(r.Context(), promptID, days)
		if err != nil {
			trend = []client.DailyTrend{}
		}

		var maxExec int64
		for _, t := range trend {
			if t.TotalExecutions > maxExec {
				maxExec = t.TotalExecutions
			}
		}

		render(w, r, templates.DailyTrendPartial(promptID, trend, maxExec))
	}
}

// GetProjectAnalyticsPartial fetches and renders project analytics data.
func (h *PartialHandler) GetProjectAnalyticsPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")

		analytics, err := h.api.GetProjectAnalytics(r.Context(), projectID)
		if err != nil {
			analytics = []client.ProjectAnalytics{}
		}

		data := templates.ProjectAnalyticsData{
			ProjectID:   projectID,
			ProjectName: "Project",
			Prompts:     analytics,
		}

		render(w, r, templates.ProjectAnalyticsPartial(data))
	}
}

// --- Close Session ---

// CloseSession handles HTMX PUT to close a consulting session.
func (h *PartialHandler) CloseSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "id")

		session, err := h.api.CloseSession(r.Context(), sessionID)
		if err != nil {
			renderSnackbar(w, r, "Error closing session: "+err.Error(), true)
			return
		}

		render(w, r, templates.ChatHeaderClosed(session))
	}
}

// --- Embedding Status ---

// GetEmbeddingStatus fetches and renders the embedding service status badge.
func (h *PartialHandler) GetEmbeddingStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := h.api.GetEmbeddingStatus(r.Context())
		if err != nil {
			status = map[string]string{"status": "error"}
		}

		render(w, r, templates.EmbeddingStatusBadge(status))
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
