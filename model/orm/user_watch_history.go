package orm

import (
	"time"
)

type UserWatchHistory struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int64     `json:"user_id" gorm:"not null;index:idx_watch_history_user;index:idx_watch_history_composite,priority:1"`
	ContentID int64     `json:"content_id" gorm:"not null;index:idx_watch_history_content"`
	WatchedAt time.Time `json:"watched_at" gorm:"not null;default:NOW();index:idx_watch_history_composite,sort:desc,priority:2"`

	Users   Users   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Content Content `gorm:"foreignKey:ContentID;constraint:OnDelete:CASCADE"`
}

func (u *UserWatchHistory) TableName() string {
	return "user_watch_history"
}
