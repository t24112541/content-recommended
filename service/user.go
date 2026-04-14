package service

import (
	"content-recommended/model/orm"
	"content-recommended/model/request"
	"content-recommended/model/response"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/fatih/color"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type redisConfig struct {
	RedisTTL int `env:"REDIS_TTL"` // min
}

type resourceLog struct {
	db        *gorm.DB
	redis     *redis.Client
	redisConf *redisConfig
}

var (
	ErrModelUnavailable = errors.New("Recommendation model is temporarily unavailable")
)

func NewUserService(db *gorm.DB, redisClient *redis.Client) *resourceLog {
	redisConf := redisConfig{}
	if err := env.Parse(&redisConf); err != nil {
		log.Printf("Failed Parse redis environment: %s", color.RedString(err.Error()))
	}

	return &resourceLog{
		db:        db,
		redis:     redisClient,
		redisConf: &redisConf,
	}
}

func (r *resourceLog) GetUsersRecommendations(req request.GetUsersRecommendations) (res response.GetUsersRecommendations, err error) {
	startedAt := time.Now().UTC()

	var totalUsers int64
	if err = r.db.Model(&orm.Users{}).Count(&totalUsers).Error; err != nil {
		return
	}

	offset := (req.Page - 1) * req.Limit
	var users []orm.Users
	if err = r.db.Model(&orm.Users{}).
		Order("id ASC").
		Offset(offset).
		Limit(req.Limit).
		Find(&users).Error; err != nil {
		return
	}

	results := make([]response.UserRecommendations, len(users))
	if len(users) > 0 {
		workerCount := req.Limit
		if workerCount > 8 {
			workerCount = 8
		}

		if workerCount > len(users) {
			workerCount = len(users)
		}

		jobs := make(chan int)
		var wg sync.WaitGroup

		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for idx := range jobs {
					userRec, userErr := r.GetUserRecommendations(request.GetUserRecommendations{
						UserId: users[idx].ID,
						Limit:  5,
					})

					if userErr != nil {
						errCode := "internal_error"
						if errors.Is(userErr, ErrModelUnavailable) {
							errCode = "model_unavailable"
						} else if errors.Is(userErr, gorm.ErrRecordNotFound) {
							errCode = "user_not_found"
						}

						results[idx] = response.UserRecommendations{
							UserID:  users[idx].ID,
							Status:  "failed",
							Error:   errCode,
							Message: userErr.Error(),
						}

						continue
					}

					results[idx] = response.UserRecommendations{
						UserID:          users[idx].ID,
						Recommendations: userRec.Recommendations,
						Status:          "success",
					}
				}
			}()
		}

		for idx := range users {
			jobs <- idx
		}

		close(jobs)
		wg.Wait()
	}

	var successCount int64
	var failedCount int64
	for _, result := range results {
		if result.Status == "success" {
			successCount++
			continue
		}
		failedCount++
	}

	res = response.GetUsersRecommendations{
		Page:       int64(req.Page),
		Limit:      int64(req.Limit),
		TotalUsers: totalUsers,
		Results:    results,
		Summary: response.GetUsersRecommendationsSummary{
			SuccessCount:     successCount,
			FailedCount:      failedCount,
			ProcessingTimeMs: time.Since(startedAt).Milliseconds(),
		},
		Metadata: response.GetUsersRecommendationsMetaData{
			GeneratedAt: time.Now().UTC(),
		},
	}

	return
}

func (r *resourceLog) GetUserRecommendations(req request.GetUserRecommendations) (res response.GetUserRecommendations, err error) {
	RedisCacheKey := fmt.Sprintf("rec:user:%d:limit:%d", req.UserId, req.Limit)
	RedisCacheTTL := time.Duration(r.redisConf.RedisTTL) * time.Minute
	RedisCacheCtx := context.Background()

	// ?: if has redis
	if r.redis != nil {
		cachedPayload, cacheErr := r.redis.Get(RedisCacheCtx, RedisCacheKey).Result()
		// ?: if found data in redis or error case not found
		if cacheErr == nil {
			if unmarshalErr := json.Unmarshal([]byte(cachedPayload), &res); unmarshalErr == nil {
				res.Metadata.CacheHit = true
				return
			}
		} else if !errors.Is(cacheErr, redis.Nil) {
			err = cacheErr
			return
		}
	}

	var user orm.Users

	err = r.db.Model(&orm.Users{}).Where("id = ?", req.UserId).First(&user).Error
	if err != nil {
		return
	}

	// ?: latency sim
	seed := time.Now().UnixNano() + req.UserId
	rng := rand.New(rand.NewSource(seed))
	time.Sleep(time.Duration(30+rng.Intn(21)) * time.Millisecond)

	if rng.Float64() < 0.015 {
		err = ErrModelUnavailable
		return
	}

	var watchedContentIDs []int64
	err = r.db.Model(&orm.UserWatchHistory{}).
		Where("user_id = ?", req.UserId).
		Pluck("content_id", &watchedContentIDs).Error
	if err != nil {
		return
	}

	var watchHistories []orm.UserWatchHistory
	err = r.db.Model(&orm.UserWatchHistory{}).
		Where("user_id = ?", req.UserId).
		Preload("Content").
		Order("watched_at DESC").
		Limit(50).
		Find(&watchHistories).Error
	if err != nil {
		return
	}

	genreCounts := make(map[string]int)
	totalWatches := 0
	for _, history := range watchHistories {
		if history.Content.Genre == "" {
			continue
		}
		genreCounts[history.Content.Genre]++
		totalWatches++
	}

	genrePreferences := make(map[string]float64, len(genreCounts))
	for genre, count := range genreCounts {
		if totalWatches == 0 {
			genrePreferences[genre] = 0
			continue
		}
		genrePreferences[genre] = float64(count) / float64(totalWatches)
	}

	var candidates []orm.Content
	query := r.db.Model(&orm.Content{})
	if len(watchedContentIDs) > 0 {
		query = query.Where("id NOT IN ?", watchedContentIDs)
	}

	err = query.Order("popularity_score DESC").Limit(100).Find(&candidates).Error
	if err != nil {
		return
	}

	type scoredRecommendation struct {
		item  response.GetUserRecommendationItem
		score float64
	}

	scored := make([]scoredRecommendation, 0, len(candidates))
	now := time.Now().UTC()
	for _, content := range candidates {
		daysSinceCreation := now.Sub(content.CreatedAt).Hours() / 24
		recencyFactor := 1.0 / (1.0 + daysSinceCreation/365.0)

		genreWeight := genrePreferences[content.Genre]
		if genreWeight == 0 {
			genreWeight = 0.1
		}

		popularityComponent := content.PopularityScore * 0.4
		genreComponent := genreWeight * 0.35
		recencyComponent := recencyFactor * 0.15
		randomNoise := (rng.Float64()*0.1 - 0.05) * 0.1
		finalScore := popularityComponent + genreComponent + recencyComponent + randomNoise

		scored = append(scored, scoredRecommendation{
			item: response.GetUserRecommendationItem{
				ContentID:       content.ID,
				Title:           content.Title,
				Genre:           content.Genre,
				PopularityScore: content.PopularityScore,
				Score:           finalScore,
			},
			score: finalScore,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	limit := req.Limit
	if limit > len(scored) {
		limit = len(scored)
	}

	// ?: get recommend with limit rec
	recommendations := make([]response.GetUserRecommendationItem, 0, limit)
	for _, recommendation := range scored[:limit] {
		recommendations = append(recommendations, recommendation.item)
	}

	res = response.GetUserRecommendations{
		UserID:          req.UserId,
		Recommendations: recommendations,
		Metadata: response.GetUserRecommendationMetaData{
			CacheHit:    false,
			GeneratedAt: now,
			TotalCount:  len(recommendations),
		},
	}

	// ?: set data to redis
	if r.redis != nil {
		serialized, marshalErr := json.Marshal(res)
		if marshalErr != nil {
			err = marshalErr
			return
		}

		if setErr := r.redis.Set(RedisCacheCtx, RedisCacheKey, serialized, RedisCacheTTL).Err(); setErr != nil {
			err = setErr
			return
		}
	}

	return
}
