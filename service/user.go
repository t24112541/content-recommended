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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/caarlos0/env/v11"
	"github.com/fatih/color"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type redisConfig struct {
	RedisTTL time.Duration `env:"REDIS_TTL" envDefault:"10"` // min
}

type resourceLog struct {
	db       *gorm.DB
	redis    *redis.Client
	redisCof *redisConfig
}

func NewUserService(db *gorm.DB, redisClient *redis.Client) *resourceLog {
	redisConf := redisConfig{}
	if err := env.Parse(&redisConf); err != nil {
		log.Printf("Failed Parse redis environment: %s", color.RedString(err.Error()))
	}

	return &resourceLog{
		db:       db,
		redis:    redisClient,
		redisCof: &redisConf,
	}
}

var ErrModelUnavailable = errors.New("Recommendation model is temporarily unavailable")

func (r *resourceLog) GetUsers(req request.GetUsers) (res []response.GetUsers, err error) {
	var users []orm.Users

	db := r.db.Model(&orm.Users{}).Find(&users)
	err = db.Error
	if err != nil {
		return
	}

	res = make([]response.GetUsers, 0, len(users))
	for _, user := range users {
		res = append(res, mapUserToResponse(user))
	}

	return
}

func (r *resourceLog) GetUserRecommendations(req request.GetUserRecommendations) (res response.GetUserRecommendations, err error) {
	cacheKey := fmt.Sprintf("rec:user:%d:limit:%d", req.UserId, req.Limit)
	cacheTTL := r.redisCof.RedisTTL * time.Minute
	cacheCtx := context.Background()

	// ?: if has redis
	if r.redis != nil {
		cachedPayload, cacheErr := r.redis.Get(cacheCtx, cacheKey).Result()
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

		if setErr := r.redis.Set(cacheCtx, cacheKey, serialized, cacheTTL).Err(); setErr != nil {
			err = setErr
			return
		}
	}

	return
}

func (r *resourceLog) GetUser(id *int64) (res orm.Users, err error) {
	db := r.db.Model(&orm.Users{})
	db = db.Where("id = ?", aws.ToInt64(id)).First(&res)
	err = db.Error

	return
}

func (r *resourceLog) CreateUsers(req orm.Users) (err error) {
	db := r.db.Create(&req)
	err = db.Error

	return
}

func mapUserToResponse(user orm.Users) response.GetUsers {
	watchHistories := make([]response.UserWatchHistory, 0, len(user.WatchHistories))
	for _, history := range user.WatchHistories {
		watchHistories = append(watchHistories, response.UserWatchHistory{
			ID:        history.ID,
			UserID:    history.UserID,
			ContentID: history.ContentID,
			WatchedAt: history.WatchedAt,
		})
	}

	return response.GetUsers{
		ID:               user.ID,
		Age:              user.Age,
		Country:          user.Country,
		SubscriptionType: user.SubscriptionType,
		CreatedAt:        user.CreatedAt,
		WatchHistories:   watchHistories,
	}
}
