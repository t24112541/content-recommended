package router

import (
	"content-recommended/handler"

	"github.com/gofiber/fiber/v2"
)

func (r *ResourceRouter) userRouter(route fiber.Router) {
	userHandler := handler.NewUserHandler(r.db, r.redis, r.validator, r.sessionStore)

	route.Get("/users/:user_id/recommendations", userHandler.GetUserRecommendations)
	route.Get("/recommendations/batch", userHandler.GetUsersRecommendations)
}
