package analytics

import (
	db "utils/db/db"
)

type promptAnalyticsResponse struct {
	VersionNumber   int32  `json:"version_number"`
	TotalExecutions int64  `json:"total_executions"`
	AvgTokens       int32  `json:"avg_tokens"`
	AvgLatencyMs    int32  `json:"avg_latency_ms"`
	TotalCost       string `json:"total_cost"`
	SuccessCount    int64  `json:"success_count"`
	ErrorCount      int64  `json:"error_count"`
}

func toPromptAnalyticsResponse(row db.GetPromptAnalyticsRow) promptAnalyticsResponse {
	return promptAnalyticsResponse{
		VersionNumber:   row.VersionNumber,
		TotalExecutions: row.TotalExecutions,
		AvgTokens:       row.AvgTokens,
		AvgLatencyMs:    row.AvgLatencyMs,
		TotalCost:       row.TotalCost,
		SuccessCount:    row.SuccessCount,
		ErrorCount:      row.ErrorCount,
	}
}

type versionAnalyticsResponse struct {
	PromptID        string `json:"prompt_id"`
	VersionNumber   int32  `json:"version_number"`
	TotalExecutions int64  `json:"total_executions"`
	AvgTokens       int32  `json:"avg_tokens"`
	AvgLatencyMs    int32  `json:"avg_latency_ms"`
	TotalCost       string `json:"total_cost"`
	AvgCost         string `json:"avg_cost"`
	SuccessCount    int64  `json:"success_count"`
	ErrorCount      int64  `json:"error_count"`
}

func toVersionAnalyticsResponse(row db.GetPromptVersionAnalyticsRow) versionAnalyticsResponse {
	return versionAnalyticsResponse{
		PromptID:        row.PromptID.String(),
		VersionNumber:   row.VersionNumber,
		TotalExecutions: row.TotalExecutions,
		AvgTokens:       row.AvgTokens,
		AvgLatencyMs:    row.AvgLatencyMs,
		TotalCost:       row.TotalCost,
		AvgCost:         row.AvgCost,
		SuccessCount:    row.SuccessCount,
		ErrorCount:      row.ErrorCount,
	}
}

type projectAnalyticsResponse struct {
	PromptID        string `json:"prompt_id"`
	PromptName      string `json:"prompt_name"`
	TotalExecutions int64  `json:"total_executions"`
	AvgTokens       int32  `json:"avg_tokens"`
	AvgLatencyMs    int32  `json:"avg_latency_ms"`
	TotalCost       string `json:"total_cost"`
}

func toProjectAnalyticsResponse(row db.GetProjectAnalyticsRow) projectAnalyticsResponse {
	return projectAnalyticsResponse{
		PromptID:        row.PromptID.String(),
		PromptName:      row.PromptName,
		TotalExecutions: row.TotalExecutions,
		AvgTokens:       row.AvgTokens,
		AvgLatencyMs:    row.AvgLatencyMs,
		TotalCost:       row.TotalCost,
	}
}

type dailyTrendResponse struct {
	Day             string `json:"day"`
	TotalExecutions int64  `json:"total_executions"`
	AvgTokens       int32  `json:"avg_tokens"`
	AvgLatencyMs    int32  `json:"avg_latency_ms"`
	TotalCost       string `json:"total_cost"`
}

func toDailyTrendResponse(row db.GetDailyTrendRow) dailyTrendResponse {
	return dailyTrendResponse{
		Day:             row.Day.Format("2006-01-02"),
		TotalExecutions: row.TotalExecutions,
		AvgTokens:       row.AvgTokens,
		AvgLatencyMs:    row.AvgLatencyMs,
		TotalCost:       row.TotalCost,
	}
}
