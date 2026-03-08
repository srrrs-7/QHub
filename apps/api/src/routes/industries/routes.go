package industries

import (
	domain "api/src/domain/consulting"
	"api/src/infra/rds/repoerr"
	"api/src/routes/response"
	"encoding/json"
	"net/http"
	"strings"
	"utils/db/db"

	"github.com/go-chi/chi/v5"
)

func (h *IndustryHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodePostIndustryConfigRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		cfg := domain.IndustryConfig{
			Slug:            req.Slug,
			Name:            req.Name,
			Description:     req.Description,
			KnowledgeBase:   req.KnowledgeBase,
			ComplianceRules: req.ComplianceRules,
		}

		created, err := h.industryRepo.Create(r.Context(), cfg)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toIndustryConfigResponse(created))
	}
}

func (h *IndustryHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		configs, err := h.industryRepo.FindAll(r.Context())
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(configs, toIndustryConfigResponse))
	}
}

func (h *IndustryHandler) GetBySlug() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")

		cfg, err := h.industryRepo.FindBySlug(r.Context(), slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toIndustryConfigResponse(cfg))
	}
}

func (h *IndustryHandler) PutBySlug() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")

		existing, err := h.industryRepo.FindBySlug(r.Context(), slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePutIndustryConfigRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		existing.Name = req.Name
		existing.Description = req.Description
		if req.KnowledgeBase != nil {
			existing.KnowledgeBase = req.KnowledgeBase
		}
		if req.ComplianceRules != nil {
			existing.ComplianceRules = req.ComplianceRules
		}

		updated, err := h.industryRepo.Update(r.Context(), existing)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toIndustryConfigResponse(updated))
	}
}

func (h *IndustryHandler) ListBenchmarks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")

		cfg, err := h.industryRepo.FindBySlug(r.Context(), slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		benchmarks, err := h.q.ListPlatformBenchmarks(r.Context(), db.ListPlatformBenchmarksParams{
			IndustryConfigID: cfg.ID,
			Limit:            20,
		})
		if err != nil {
			response.HandleError(w, repoerr.Handle(err, "IndustryHandler", ""))
			return
		}

		response.OK(w, response.MapSlice(benchmarks, toBenchmarkResponse))
	}
}

func (h *IndustryHandler) ComplianceCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")

		cfg, err := h.industryRepo.FindBySlug(r.Context(), slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodeComplianceCheckRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		violations := checkCompliance(cfg.ComplianceRules, req.Content)

		response.OK(w, complianceCheckResponse{
			Compliant:  len(violations) == 0,
			Violations: violations,
		})
	}
}

// checkCompliance performs a simple rule-based compliance check.
// ComplianceRules JSON format: {"rules": [{"keyword": "...", "message": "..."}]}
func checkCompliance(rulesJSON json.RawMessage, content string) []complianceIssue {
	if rulesJSON == nil {
		return nil
	}

	var parsed struct {
		Rules []struct {
			Keyword string `json:"keyword"`
			Message string `json:"message"`
		} `json:"rules"`
	}

	if err := json.Unmarshal(rulesJSON, &parsed); err != nil {
		return nil
	}

	lower := strings.ToLower(content)
	var violations []complianceIssue
	for _, rule := range parsed.Rules {
		if rule.Keyword != "" && strings.Contains(lower, strings.ToLower(rule.Keyword)) {
			violations = append(violations, complianceIssue{
				Rule:    rule.Keyword,
				Message: rule.Message,
			})
		}
	}
	return violations
}
