package postgres

import (
	"content-recommended/model/orm"
	"fmt"
	"math"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

const (
	seedValue           int64 = 20260413
	minimumUsers              = 20
	minimumContentItems       = 50
	minimumWatchHistory       = 200
)

var (
	seedCountries = []string{"US", "TH", "JP", "DE", "BR", "IN", "GB", "CA", "AU", "FR"}
	seedGenres    = []string{"action", "drama", "comedy", "thriller", "documentary", "romance", "sci-fi"}
)

func (r *resourceDB) seedData() (err error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			tx.Rollback()
			panic(recovered)
		}
	}()

	rng := rand.New(rand.NewSource(seedValue))

	if err = seedUsers(tx, rng); err != nil {
		tx.Rollback()
		return
	}

	if err = seedContent(tx, rng); err != nil {
		tx.Rollback()
		return
	}

	if err = seedWatchHistory(tx, rng); err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit().Error
	return
}

func seedUsers(tx *gorm.DB, rng *rand.Rand) (err error) {
	var existingCount int64
	if err = tx.Model(&orm.Users{}).Count(&existingCount).Error; err != nil {
		return
	}

	remaining := minimumUsers - int(existingCount)
	if remaining <= 0 {
		return nil
	}

	users := make([]orm.Users, 0, remaining)
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < remaining; i++ {
		index := int(existingCount) + i + 1
		users = append(users, orm.Users{
			Age:              (18 + rng.Intn(53)),
			Country:          pickCountry(index),
			SubscriptionType: pickSubscription(rng),
			CreatedAt:        baseTime.Add(time.Duration(index*6) * time.Hour),
		})
	}

	err = tx.CreateInBatches(users, 100).Error
	return
}

func seedContent(tx *gorm.DB, rng *rand.Rand) (err error) {
	var existingCount int64
	if err = tx.Model(&orm.Content{}).Count(&existingCount).Error; err != nil {
		return
	}

	remaining := minimumContentItems - int(existingCount)
	if remaining <= 0 {
		return nil
	}

	items := make([]orm.Content, 0, remaining)
	baseTime := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < remaining; i++ {
		rank := float64(int(existingCount) + i + 1)
		items = append(items, orm.Content{
			Title:           fmt.Sprintf("Content %03d", int(existingCount)+i+1),
			Genre:           seedGenres[(int(existingCount)+i)%len(seedGenres)],
			PopularityScore: powerLawScore(rank, rng),
			CreatedAt:       baseTime.Add(time.Duration(i*4) * time.Hour),
		})
	}

	err = tx.CreateInBatches(items, 100).Error
	return
}

func seedWatchHistory(tx *gorm.DB, rng *rand.Rand) (err error) {
	var existingCount int64
	if err = tx.Model(&orm.UserWatchHistory{}).Count(&existingCount).Error; err != nil {
		return
	}

	remaining := minimumWatchHistory - int(existingCount)
	if remaining <= 0 {
		return nil
	}

	var userIDs []int64
	if err = tx.Model(&orm.Users{}).Order("id ASC").Pluck("id", &userIDs).Error; err != nil {
		return
	}

	var contentIDs []int64
	if err = tx.Model(&orm.Content{}).Order("id ASC").Pluck("id", &contentIDs).Error; err != nil {
		return
	}

	if len(userIDs) == 0 || len(contentIDs) == 0 {
		return nil
	}

	history := make([]orm.UserWatchHistory, 0, remaining)
	baseTime := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	contentZipf := rand.NewZipf(rng, 1.25, 1, uint64(len(contentIDs)-1))

	for i := 0; i < remaining; i++ {
		userID := userIDs[rng.Intn(len(userIDs))]
		contentID := contentIDs[int(contentZipf.Uint64())]

		history = append(history, orm.UserWatchHistory{
			UserID:    userID,
			ContentID: contentID,
			WatchedAt: baseTime.Add(time.Duration(int(existingCount)+i) * time.Hour),
		})
	}

	err = tx.CreateInBatches(history, 200).Error
	return
}

func pickCountry(index int) string {
	return seedCountries[(index-1)%len(seedCountries)]
}

func pickSubscription(rng *rand.Rand) string {
	roll := rng.Float64()
	if roll < 0.62 {
		return "free"
	}
	if roll < 0.87 {
		return "basic"
	}

	return "premium"
}

func powerLawScore(rank float64, rng *rand.Rand) float64 {
	base := 1000.0 / math.Pow(rank, 1.18)
	noise := rng.Float64() * 2.5

	return base + noise
}
