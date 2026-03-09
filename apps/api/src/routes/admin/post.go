package admin

import (
	"api/src/routes/response"
	"net/http"
)

// PostAggregate triggers a monthly batch aggregation of execution log metrics.
func (h *AdminHandler) PostAggregate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := h.batch.RunMonthlyAggregation(r.Context())
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toAggregationResponse(result))
	}
}
