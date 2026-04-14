package response

import (
	"time"
)

type GetUserRecommendations struct {
	UserID          int64                         `json:"user_id"`
	Recommendations []GetUserRecommendationItem   `json:"recommendations"`
	Metadata        GetUserRecommendationMetaData `json:"metadata"`
}

type GetUserRecommendationItem struct {
	ContentID       int64   `json:"content_id"`
	Title           string  `json:"title"`
	Genre           string  `json:"genre"`
	PopularityScore float64 `json:"popularity_score"`
	Score           float64 `json:"score"`
}

type GetUserRecommendationMetaData struct {
	CacheHit    bool      `json:"cache_hit"`
	GeneratedAt time.Time `json:"generated_at"`
	TotalCount  int       `json:"total_count"`
}
