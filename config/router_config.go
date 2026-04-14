package config

import (
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/gorm"
)

type ConfigRouter struct {
	db       *gorm.DB
	validate *validator.Validate
	appMode  string `env:"APP_MODE"`
}

func NewConfigRouter(
	db *gorm.DB,
	validate *validator.Validate,
) *ConfigRouter {
	var (
		conf ConfigRouter
	)

	_ = env.Parse(&conf)
	return &ConfigRouter{
		db:       db,
		validate: validate,
		appMode:  conf.appMode,
	}
}

func (r *ConfigRouter) Config(app *fiber.App) *fiber.App {
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${client-id} | ${ip} | ${x-forwarded-for} | ${method} | ${path}\n",
		TimeFormat: "2006/01/02 15:04:05",
		Output:     os.Stdout,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:               200000, // TODO: adjust
		Expiration:        60 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))

	return app
}
