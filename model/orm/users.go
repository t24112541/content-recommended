package orm

import (
	"time"
)

type Users struct {
	ID               int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Age              int       `json:"age" gorm:"type:int;not null;check:age > 0"`
	Country          string    `json:"country" gorm:"type:varchar(2);not null;index:idx_users_country"`
	SubscriptionType string    `json:"subscription_type" gorm:"type:varchar(20);not null;index:idx_users_subscription"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`

	WatchHistories []UserWatchHistory `json:"watch_histories" gorm:"foreignKey:UserID"`
}

func (s *Users) TableName() string {
	return "users"
}
