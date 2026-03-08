package industries

import (
	"api/src/routes/requtil"
	"encoding/json"
	"net/http"
)

type postIndustryConfigRequest struct {
	Slug            string          `json:"slug" validate:"required,min=2,max=80"`
	Name            string          `json:"name" validate:"required,min=1,max=200"`
	Description     string          `json:"description" validate:"omitempty,max=1000"`
	KnowledgeBase   json.RawMessage `json:"knowledge_base"`
	ComplianceRules json.RawMessage `json:"compliance_rules"`
}

func decodePostIndustryConfigRequest(r *http.Request) (postIndustryConfigRequest, error) {
	return requtil.Decode(r, func(req *postIndustryConfigRequest) {
		req.Slug = requtil.Sanitize.Sanitize(req.Slug)
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Description = requtil.Sanitize.Sanitize(req.Description)
	})
}

type putIndustryConfigRequest struct {
	Name            string          `json:"name" validate:"omitempty,min=1,max=200"`
	Description     string          `json:"description" validate:"omitempty,max=1000"`
	KnowledgeBase   json.RawMessage `json:"knowledge_base"`
	ComplianceRules json.RawMessage `json:"compliance_rules"`
}

func decodePutIndustryConfigRequest(r *http.Request) (putIndustryConfigRequest, error) {
	return requtil.Decode(r, func(req *putIndustryConfigRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Description = requtil.Sanitize.Sanitize(req.Description)
	})
}

type complianceCheckRequest struct {
	Content string `json:"content" validate:"required,min=1"`
}

func decodeComplianceCheckRequest(r *http.Request) (complianceCheckRequest, error) {
	return requtil.Decode(r, func(req *complianceCheckRequest) {
		req.Content = requtil.Sanitize.Sanitize(req.Content)
	})
}
