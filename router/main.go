package router

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ResourceRouter struct {
	sessionStore *session.Store
	db           *gorm.DB
	redis        *redis.Client
	validator    *validator.Validate
}

func New(
	sessionStore *session.Store,
	db *gorm.DB,
	redisClient *redis.Client,
	validator *validator.Validate,
) *ResourceRouter {
	return &ResourceRouter{
		sessionStore: sessionStore,
		db:           db,
		redis:        redisClient,
		validator:    validator,
	}
}

// Routes for fiber
func (r *ResourceRouter) Init(app *fiber.App) {

	app.Route("/", r.userRouter)

}
