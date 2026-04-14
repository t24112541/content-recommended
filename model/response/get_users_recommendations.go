package response

import "time"

type GetUsersRecommendations struct {
	Page       int64                           `json:"page"`
	Limit      int64                           `json:"limit"`
	TotalUsers int64                           `json:"total_users"`
	Results    []UserRecommendations           `json:"results"`
	Summary    GetUsersRecommendationsSummary  `json:"summary"`
	Metadata   GetUsersRecommendationsMetaData `json:"metadata"`
}

type UserRecommendations struct {
	UserID          int64                       `json:"user_id"`
	Recommendations []GetUserRecommendationItem `json:"recommendations"`
	Status          string                      `json:"status"`
	Error           string                      `json:"error,omitempty"`
	Message         string                      `json:"message,omitempty"`
}

type GetUsersRecommendationsSummary struct {
	SuccessCount     int64 `json:"success_count"`
	FailedCount      int64 `json:"failed_count"`
	ProcessingTimeMs int64 `json:"processing_time_ms"`
}

type GetUsersRecommendationsMetaData struct {
	GeneratedAt time.Time `json:"generated_at"`
}
