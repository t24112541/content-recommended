package orm

import (
	"time"
)

type Content struct {
	ID              int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Title           string    `json:"title" gorm:"type:varchar(255);not null"`
	Genre           string    `json:"genre" gorm:"type:varchar(20);not null;index:idx_content_genre"`
	PopularityScore float64   `json:"popularity_score" gorm:"type:double precision;not null;index:idx_content_popularity,sort:desc"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`

	WatchHistories []UserWatchHistory `gorm:"foreignKey:ContentID"`
}

func (s *Content) TableName() string {
	return "content"
}
