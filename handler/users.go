package handler

import (
	"content-recommended/model/request"
	"content-recommended/model/response"
	"content-recommended/service"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserResource struct {
	db           *gorm.DB
	redis        *redis.Client
	validator    *validator.Validate
	sessionStore *session.Store
}

func NewUserHandler(
	db *gorm.DB,
	redisClient *redis.Client,
	validator *validator.Validate,
	sessionStore *session.Store,
) *UserResource {
	return &UserResource{
		db:           db,
		redis:        redisClient,
		validator:    validator,
		sessionStore: sessionStore,
	}
}

func (r *UserResource) GetUsersRecommendations(ctx *fiber.Ctx) (err error) {
	var (
		req request.GetUsersRecommendations
	)

	if err := ctx.QueryParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(response.ErrorResponse{
			Error:   "invalid_parameter",
			Message: "Invalid limit parameter",
		})
	}

	if req.Page < 1 {
		req.Page = 1
	}

	if req.Limit < 1 {
		req.Limit = 20
	}

	if req.Limit > 100 {
		req.Limit = 100
	}

	if err = r.validator.Struct(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(response.ErrorResponse{
			Error:   "invalid_parameter",
			Message: err.Error(),
		})
	}

	userService := service.NewUserService(r.db, r.redis)
	data, err := userService.GetUsersRecommendations(req)
	if err != nil {
		if errors.Is(err, service.ErrModelUnavailable) {
			return ctx.Status(http.StatusServiceUnavailable).JSON(response.ErrorResponse{
				Error:   "model_unavailable",
				Message: err.Error(),
			})
		}

		return ctx.Status(http.StatusInternalServerError).JSON(response.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(data)

}

func (r *UserResource) GetUserRecommendations(ctx *fiber.Ctx) (err error) {
	var (
		req request.GetUserRecommendations
	)

	req.Limit = 10

	if err := ctx.ParamsParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(response.ErrorResponse{
			Error:   "invalid_parameter",
			Message: err.Error(),
		})
	}

	if err := ctx.QueryParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(response.ErrorResponse{
			Error:   "invalid_parameter",
			Message: "Invalid limit parameter",
		})
	}

	if req.Limit < 1 || req.Limit > 50 {
		return ctx.Status(http.StatusBadRequest).JSON(response.ErrorResponse{
			Error:   "invalid_parameter",
			Message: "Invalid limit parameter",
		})
	}

	if err = r.validator.Struct(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(response.ErrorResponse{
			Error:   "invalid_parameter",
			Message: err.Error(),
		})
	}

	userService := service.NewUserService(r.db, r.redis)
	data, err := userService.GetUserRecommendations(req)
	if err != nil {
		if errors.Is(err, service.ErrModelUnavailable) {
			return ctx.Status(http.StatusServiceUnavailable).JSON(response.ErrorResponse{
				Error:   "model_unavailable",
				Message: err.Error(),
			})
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(response.ErrorResponse{
				Error:   "user_not_found",
				Message: fmt.Sprintf("User with ID %d does not exist", req.UserId),
			})
		}

		return ctx.Status(http.StatusInternalServerError).JSON(response.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(data)

}
